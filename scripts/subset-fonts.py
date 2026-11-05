#!/usr/bin/env python3
"""Build pinned font subsets. Requires fonttools==4.60.2; never used at runtime.
Usage: python scripts/subset-fonts.py CACHE [--download] [--cjk | --packs | --emoji]
The download step verifies sources.json hashes before instancing/subsetting.
--emoji copies the CBDT font into packs/emoji and derives packs/emoji/sbix.
"""
import argparse, hashlib, json, shutil, urllib.request, zipfile, tempfile
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
mode=parser.add_mutually_exclusive_group()
mode.add_argument('--cjk',action='store_true',help='build only the optional CJK module')
mode.add_argument('--packs',action='store_true',help='build only the optional script packs module')
mode.add_argument('--emoji',action='store_true',help='copy the optional emoji module font and derive its sbix companion')
args=parser.parse_args()
root=Path(__file__).resolve().parents[1]
args.cache.mkdir(parents=True,exist_ok=True)
output=root/'cjk' if args.cjk else root/'packs' if args.packs else root/'packs'/'emoji' if args.emoji else root
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
# Text subsets keep composition, positioning and kerning only; nttypeset
# disables ligatures at runtime. Script packs keep every layout feature
# because Arabic joining and Indic reordering live in GSUB/GPOS.
TEXT_FEATURES=['ccmp','locl','mark','mkmk','kern']
PACKS=[('arabic','Arabic',[(0x600,0x6ff),(0x750,0x77f),(0x8a0,0x8ff),(0xfb50,0xfdff),(0xfe70,0xfeff)]),
 ('hebrew','Hebrew',[(0x590,0x5ff),(0xfb1d,0xfb4f)]),
 ('devanagari','Devanagari',[(0x900,0x97f),(0x1cd0,0x1cff),(0xa8e0,0xa8ff)]),
 ('bengali','Bengali',[(0x980,0x9ff)]),
 ('tamil','Tamil',[(0xb80,0xbff),(0x11fc0,0x11fff)]),
 ('thai','Thai',[(0xe00,0xe7f)]),
 ('khmer','Khmer',[(0x1780,0x17ff),(0x19e0,0x19ff)])]
# Space, no-break space, joiners, directional controls and the replacement
# character stay in every pack face so shaping never falls out of it.
CONTROLS={0x20,0xa0,0x200b,0x200c,0x200d,0x200e,0x200f,0x2066,0x2067,0x2068,0x2069,0xfffd}


def instance(font, weight, optical=None):
 if 'fvar' not in font:return font
 settings={a.axisTag:a.defaultValue for a in font['fvar'].axes};settings['wght']=weight
 if optical is not None:settings['opsz']=optical
 return instantiateVariableFont(font,settings,inplace=True)


def save(font, codepoints, ident, variant, dest, features=TEXT_FEATURES, prefix='Gloss Text '):
 options=subset.Options();options.hinting=False
 options.layout_features=features
 options.name_IDs=['*'];options.name_legacy=True;options.name_languages=['*']
 sub=subset.Subsetter(options=options);sub.populate(unicodes=codepoints);sub.subset(font)
 familyname=prefix+ident
 names={1:familyname,2:variant,3:'GlossSubset-'+ident+'-'+variant,4:familyname+' '+variant,6:familyname.replace(' ','')+'-'+variant,16:familyname,17:variant,21:familyname,22:variant}
 for record in font['name'].names:
  if record.nameID in names:record.string=names[record.nameID].encode(record.getEncoding(),errors='replace')
 if 'CFF ' in font:
  cff=font['CFF '].cff; cff.fontNames=[names[6]]
  cff.topDictIndex[0].FamilyName=familyname; cff.topDictIndex[0].FullName=names[4]
 font.recalcTimestamp=False
 font['head'].created=font['head'].modified=2082844800
 dest.parent.mkdir(parents=True,exist_ok=True)
 font.save(dest,reorderTables=True)
 record_hash(dest)


def record_hash(dest):
 name=dest.relative_to(output).as_posix() if dest.is_relative_to(output) else dest.name
 hashes[name]=hashlib.sha256(dest.read_bytes()).hexdigest()

hashes={}
if args.cjk:
 han={int(line,16) for line in (output/'data/gb2312-han.txt').read_text().splitlines() if line and not line.startswith('#')}
 assert len(han)==6763
 # 漢 is an explicit extra for mixed-language ruby examples; not in GB 2312.
 codepoints=han|{ord('漢')}|points([(0x20,0x7e),(0x1100,0x11ff),(0x3000,0x30ff),(0x3130,0x318f),(0xa960,0xa97f),(0xac00,0xd7ff),(0xff00,0xffef)])
 for variant in ['Regular','Bold']:
  font=TTFont(args.cache/('NotoSansCJKsc-'+variant+'.otf'),recalcTimestamp=False)
  save(font,codepoints,'CJKSC',variant,output/('sc'+variant.lower())/('GlossCJKSC-'+variant+'.otf'))
elif args.packs:
 coverage={}
 for slug,script,ranges in PACKS:
  codepoints=points(ranges)|CONTROLS
  for variant,weight in [('Regular',400),('Bold',700)]:
   font=TTFont(args.cache/('notosans'+slug+'-NotoSans'+script+'[wdth,wght].ttf'),recalcTimestamp=False)
   font=instance(font,weight)
   coverage[script+'/'+variant]=''.join(chr(cp) for cp in sorted(codepoints & set(font.getBestCmap())))
   # Public file names: hosts serve these URLs; only the name table says Gloss.
   save(font,codepoints,script,variant,output/slug/('NotoSans'+script+'-'+variant+'.ttf'),features=['*'],prefix='Gloss ')
  notice=output/'internal'/'notice'/('notosans'+slug+'.OFL.txt');notice.parent.mkdir(parents=True,exist_ok=True)
  shutil.copyfile(args.cache/('notosans'+slug+'-OFL.txt'),notice)
 (output/'coverage.json').write_text(json.dumps(coverage,ensure_ascii=False,indent=2)+'\n')
elif args.emoji:
 # The CBDT font is embedded unmodified: no subsetting, no rename, no fontTools.
 shutil.copyfile(args.cache/'NotoColorEmoji.ttf',output/'NotoColorEmoji.ttf')
 shutil.copyfile(args.cache/'notocoloremoji-OFL.txt',output/'notocoloremoji.OFL.txt')
 record_hash(output/'NotoColorEmoji.ttf')
 # The sbix companion carries the same PNG strikes for WebKit, which rejects
 # CBDT. Chrome draws sbix only for glyphs with a glyf bounding box, so each
 # bitmap glyph gets two coincident rectangles of opposite winding: a box
 # with zero fill that reads as tofu wherever sbix is unsupported. When a
 # glyph has contours, CoreText and FreeType place the bitmap relative to the
 # lower-left corner of that box, not the glyph origin (Apple Color Emoji
 # ships zero offsets over a box at the baseline), so the box is the bitmap's
 # own rectangle and the origin offsets are measured from it: zero here.
 from fontTools.ttLib import newTable
 from fontTools.ttLib.tables._g_l_y_f import Glyph
 from fontTools.ttLib.tables.sbixStrike import Strike
 from fontTools.ttLib.tables.sbixGlyph import Glyph as SbixGlyph
 from fontTools.pens.ttGlyphPen import TTGlyphPen
 sbixdir=output/'sbix'; sbixdir.mkdir(exist_ok=True)
 font=TTFont(args.cache/'NotoColorEmoji.ttf',recalcTimestamp=False)
 ppem=font['CBLC'].strikes[0].bitmapSizeTable.ppemX; upem=font['head'].unitsPerEm
 bitmaps=font['CBDT'].strikeData[0]
 table=newTable('sbix'); table.version=1; table.flags=1; table.strikes={}
 strike=Strike(ppem=ppem,resolution=72)
 glyf=newTable('glyf'); glyf.glyphs={}; glyf.glyphOrder=font.getGlyphOrder()
 for name in glyf.glyphOrder:
  if name not in bitmaps:
   glyf.glyphs[name]=Glyph(); continue
  m=bitmaps[name].metrics
  x0,y0=round(m.BearingX*upem/ppem),round((m.BearingY-m.height)*upem/ppem)
  x1,y1=round((m.BearingX+m.width)*upem/ppem),round(m.BearingY*upem/ppem)
  px=lambda units:round(units*ppem/upem)
  strike.glyphs[name]=SbixGlyph(glyphName=name,graphicType='png ',imageData=bitmaps[name].imageData,originOffsetX=m.BearingX-px(x0),originOffsetY=m.BearingY-m.height-px(y0))
  pen=TTGlyphPen(None)
  pen.moveTo((x0,y0));pen.lineTo((x0,y1));pen.lineTo((x1,y1));pen.lineTo((x1,y0));pen.closePath()
  pen.moveTo((x0,y0));pen.lineTo((x1,y0));pen.lineTo((x1,y1));pen.lineTo((x0,y1));pen.closePath()
  glyf.glyphs[name]=pen.glyph()
 table.strikes[ppem]=strike
 font['sbix']=table; font['glyf']=glyf; font['loca']=newTable('loca')
 del font['CBDT'],font['CBLC']
 version=next(r.toUnicode() for r in font['name'].names if r.nameID==5)
 names={1:'Noto Color Emoji sbix',3:'Noto Color Emoji sbix; ntfonts conversion of '+version,4:'Noto Color Emoji sbix',6:'NotoColorEmoji-sbix',10:'Color emoji font using sbix glyph data, converted by ntfonts from the Noto Color Emoji CBDT build.'}
 for record in font['name'].names:
  if record.nameID in names:record.string=names[record.nameID].encode(record.getEncoding(),errors='replace')
 dest=sbixdir/'NotoColorEmoji-sbix.ttf'; font.save(dest)
 shutil.copyfile(args.cache/'notocoloremoji-OFL.txt',sbixdir/'notocoloremoji.OFL.txt')
 (sbixdir/'subsets.sha256').write_text(f'{hashlib.sha256(dest.read_bytes()).hexdigest()}  {dest.name}\n')
 print('sbix bytes:',dest.stat().st_size)
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
   font=instance(font,weight,optical)
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
