#!/usr/bin/env python3
"""Build pinned font subsets. Requires fonttools==4.60.2; never used at runtime.
Usage: python scripts/subset-fonts.py CACHE [--download]
The download step verifies sources.json hashes before instancing/subsetting.
"""
import argparse, hashlib, json, urllib.request, zipfile, tempfile
from fontTools.merge import Merger
from pathlib import Path
from fontTools import subset
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont
import fontTools

assert fontTools.__version__ == '4.60.2', 'install fonttools==4.60.2'
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('cache',type=Path)
parser.add_argument('--download',action='store_true')
parser.add_argument('--cjk',action='store_true',help='build only the optional CJK module')
args=parser.parse_args()
root=Path(__file__).resolve().parents[1]
args.cache.mkdir(parents=True,exist_ok=True)
output=root/'cjk' if args.cjk else root
for source in json.loads((output/'sources.json').read_text()):
 path=args.cache/source['file']
 if not path.exists() and args.download:
  with urllib.request.urlopen(source['url'],timeout=120) as response:path.write_bytes(response.read())
 data=path.read_bytes()
 assert hashlib.sha256(data).hexdigest()==source['sha256'], f'checksum mismatch: {path}'
 if source['file']=='iosevka.zip':
  with zipfile.ZipFile(path) as archive:
   for variant in ['Regular','Bold','Italic','BoldItalic']:
    name='Iosevka-'+variant+'.ttf';(args.cache/name).write_bytes(archive.read(name))
# Keep only glyphs present upstream within the requested coverage ranges.
def points(ranges):
 return {cp for first,last in ranges for cp in range(first,last+1)}
unicodes=points([(0x20,0x7e),(0xa0,0x24f),(0x2b0,0x52f),(0x1e00,0x1eff),
 (0x2000,0x206f),(0x20a0,0x20cf),(0x2100,0x22ff),(0x2500,0x25ff)])|{0x2713,0x2714,0x2717,0x2718,0xfffd}
symbols=points([(0x2190,0x22ff),(0x2500,0x27bf)])|{0x20,0xfffd}


def save(font, codepoints, ident, variant, dest):
 options=subset.Options();options.hinting=False
 options.layout_features=['ccmp','locl','mark','mkmk','kern']
 options.name_IDs=['*'];options.name_legacy=True;options.name_languages=['*']
 sub=subset.Subsetter(options=options);sub.populate(unicodes=codepoints);sub.subset(font)
 familyname='Gloss Text '+ident
 names={1:familyname,2:variant,3:'GlossSubset-'+ident+'-'+variant,4:familyname+' '+variant,6:'GlossText'+ident+'-'+variant,16:familyname,17:variant,21:familyname,22:variant}
 for record in font['name'].names:
  if record.nameID in names:record.string=names[record.nameID].encode(record.getEncoding(),errors='replace')
 if 'CFF ' in font:
  cff=font['CFF '].cff; cff.fontNames=[names[6]]
  cff.topDictIndex[0].FamilyName=familyname; cff.topDictIndex[0].FullName=names[4]
 font.recalcTimestamp=False
 font['head'].created=font['head'].modified=2082844800
 font.save(dest,reorderTables=True)
 hashes[dest.name]=hashlib.sha256(dest.read_bytes()).hexdigest()

hashes={}
if args.cjk:
 han={int(line,16) for line in (output/'data/gb2312-han.txt').read_text().splitlines() if line and not line.startswith('#')}
 assert len(han)==6763
 # 漢 is an explicit extra for mixed-language ruby examples; not in GB 2312.
 codepoints=han|{ord('漢')}|points([(0x20,0x7e),(0x1100,0x11ff),(0x3000,0x30ff),(0x3130,0x318f),(0xa960,0xa97f),(0xac00,0xd7ff),(0xff00,0xffef)])
 for variant in ['Regular','Bold']:
  font=TTFont(args.cache/('NotoSansCJKsc-'+variant+'.otf'),recalcTimestamp=False)
  save(font,codepoints,'CJKSC',variant,output/('GlossCJKSC-'+variant+'.otf'))
else:
 families=[('A','literata','Literata','opsz,wght',12),('B','newsreader','Newsreader','opsz,wght',16),('C','ibmplexsans','IBMPlexSans','wdth,wght',None),('D','sourceserif4','SourceSerif4','opsz,wght',16),('E','iosevka','Iosevka',None,None),('F','ibmplexmono','IBMPlexMono',None,None),('G','jetbrainsmono','JetBrainsMono','wght',None)]
 coverage={}
 for ident,folder,family,axes,optical in families:
  for variant,weight,italic in [('Regular',400,False),('Bold',700,False),('Italic',400,True),('BoldItalic',700,True)]:
   if axes:
    name=folder+'-'+family+('-Italic' if italic else '')+'['+axes+'].ttf'
   elif folder=='iosevka':name=family+'-'+variant+'.ttf'
   else:name=folder+'-'+family+'-'+variant+'.ttf'
   font=TTFont(args.cache/name,recalcTimestamp=False)
   if 'fvar' in font:
    settings={a.axisTag:a.defaultValue for a in font['fvar'].axes};settings['wght']=weight
    if optical is not None:settings['opsz']=optical
    font=instantiateVariableFont(font,settings,inplace=True)
   coverage[family+'/'+variant]=''.join(chr(cp) for cp in sorted(unicodes & set(font.getBestCmap())))
   save(font,unicodes,ident,variant,root/('Gloss'+ident+'-'+variant+'.ttf'))
 font=TTFont(args.cache/'NotoSansSymbols2-Regular.ttf',recalcTimestamp=False)
 # Supplement only missing arrows/math from Noto Sans Math, keeping one face.
 missing=(symbols-set(font.getBestCmap())) & points([(0x2190,0x22ff)])
 math=TTFont(args.cache/'NotoSansMath-Regular.ttf',recalcTimestamp=False)
 # This is a text glyph fallback, not a mathematical layout font.
 if 'MATH' in math: del math['MATH']
 # The supplement has no vertical metrics; this fallback is horizontal only.
 for tag in ['vhea','vmtx']:
  if tag in font: del font[tag]
 with tempfile.TemporaryDirectory() as temp:
  first=Path(temp)/'symbols.ttf'; second=Path(temp)/'math.ttf'
  save(font,symbols,'Symbols','Regular',first)
  save(math,missing,'Symbols','Regular',second)
  font=Merger().merge([str(first),str(second)])
  del hashes[first.name],hashes[second.name]
  save(font,symbols,'Symbols','Regular',root/'GlossSymbols-Regular.ttf')
 (root/'coverage.json').write_text(json.dumps(coverage,ensure_ascii=False,indent=2)+'\n')
(output/'subsets.sha256').write_text(''.join(f'{digest}  {name}\n' for name,digest in sorted(hashes.items())))
print('Subset bytes:',sum((output/n).stat().st_size for n in hashes))
