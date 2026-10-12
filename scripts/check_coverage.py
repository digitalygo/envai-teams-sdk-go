import pathlib,sys
lines=pathlib.Path(sys.argv[1]).read_text().splitlines()[1:]
covered=total=0
for line in lines:
 _,count,hits=line.rsplit(' ',2)
 total+=int(count)
 if int(hits)>0:covered+=int(count)
percentage=covered*100/total if total else 0
print(f'Statement coverage: {percentage:.2f}% ({covered}/{total})')
if percentage<80:raise SystemExit('Coverage below 80%')
