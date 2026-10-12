import hashlib,json,pathlib
root=pathlib.Path(__file__).resolve().parent.parent
provenance=json.loads((root/'contract/provenance.json').read_text())
for name in ['openapi.yaml','events.json']:
 assert hashlib.sha256((root/'contract'/name).read_bytes()).hexdigest()==provenance['files'][name],name
actual=json.loads((root/'contract/operations.json').read_text());expected=json.loads((root/'testdata/operations.json').read_text())
assert len(actual)==52
assert [{k:op[k] for k in ['id','method','path','authenticated']} for op in actual]==expected
print('Exact contract hashes and 52-operation inventory verified')
