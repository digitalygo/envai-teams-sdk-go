import json,re,pathlib,yaml
root=pathlib.Path(__file__).resolve().parent.parent
class Loader(yaml.SafeLoader):pass
Loader.add_constructor('tag:yaml.org,2002:int',lambda loader,node:node.value if ':' in node.value else yaml.SafeLoader.construct_yaml_int(loader,node))
api=yaml.load((root/'contract/openapi.yaml').read_text(),Loader=Loader)
schemas=api['components']['schemas']
def name(x):
 x=re.sub(r'[^A-Za-z0-9]+',' ',x)
 s=''.join(p[0].upper()+p[1:] for p in x.split())
 for a,b in [('Id','ID'),('Url','URL'),('Jwks','JWKS'),('Jwk','JWK'),('Api','API'),('Pdf','PDF'),('Llm','LLM'),('Usd','USD')]:s=s.replace(a,b)
 if s=='Error':s='ErrorEnvelope'
 return s
extra={}
def typ(s,hint):
 if '$ref' in s:return ('ATP' if '/$defs/' in s['$ref'] else '')+name(s['$ref'].rsplit('/',1)[-1])
 t=s.get('type')
 if isinstance(t,list): t=next((x for x in t if x!='null'),None)
 if s.get('additionalProperties') is True: return 'map[string]json.RawMessage'
 if 'properties' in s:
  extra.setdefault(name(hint),s);return name(hint)
 if t=='object':return 'map[string]json.RawMessage'
 if t=='array':return '[]'+typ(s.get('items',{}),hint+'Item')
 if t=='string':return 'string'
 if t=='integer':return 'int64'
 if t=='number':return 'float64'
 if t=='boolean':return 'bool'
 return 'json.RawMessage'
def decl(n,s):
 if 'properties' not in s or s.get('additionalProperties') is True:return 'type '+n+' '+typ(s,n)+'\n'
 fields=[]
 for k,v in s['properties'].items():
  ty=typ(v,n+name(k)); required=k in s.get('required',[]); nullable='null' in v.get('type',[]) or any(x.get('type')=='null' for x in v.get('anyOf',[]))
  if not required:ty='Optional['+ty+']'; tag=k+',omitzero'
  else:
   tag=k
   if nullable and ty!='json.RawMessage' and not ty.startswith(('[]','map[')):ty='*'+ty
  fields.append('\t'+name(k)+' '+ty+' `json:"'+tag+'"`')
 return 'type '+n+' struct {\n'+'\n'.join(fields)+'\n}\n'
models=[]
for k,v in schemas.items(): models.append(decl(name(k),v))
ops=[]; wrappers=[]
for path,items in api['paths'].items():
 for method,op in items.items():
  if not isinstance(op,dict) or 'operationId' not in op:continue
  oid=op['operationId']; n=name(oid); pars=[]
  for pp in items.get('parameters',[])+op.get('parameters',[]):
   p=api['components']['parameters'][pp['$ref'].rsplit('/',1)[-1]] if '$ref' in pp else pp
   pars.append(p)
  auth=op.get('security',api.get('security',[]))!=[]
  content=op.get('requestBody',{}).get('content',{});bodytype='';bodykind=''
  if 'application/json' in content:bodytype=typ(content['application/json']['schema'],n+'Body');bodykind='json'
  elif 'multipart/form-data' in content:bodytype='[]MultipartPart';bodykind='multipart'
  response=None
  for code,r in op.get('responses',{}).items():
   if str(code).startswith('2'):
    response=r;break
  rc=(response or {}).get('content',{})
  kind='json';rt='json.RawMessage';unwrap=False
  if oid=='streamChatEvents':kind='sse';rt='*EventStream'
  elif not rc:kind='empty';rt=''
  elif 'application/json' not in rc:kind='binary';rt='*Download'
  else:
   rs=rc['application/json']['schema'];resolved=schemas.get(rs.get('$ref','').rsplit('/',1)[-1],rs)
   if set(resolved.get('properties',{}))=={'data'}:rs=resolved['properties']['data'];unwrap=True
   rt=typ(rs,n+'Result')
  ops.append({'id':oid,'method':method.upper(),'path':path,'authenticated':auth,'kind':kind,'body':bodykind,'unwrap':unwrap,'parameters':[{'name':p['name'],'in':p['in'],'required':p.get('required',False)} for p in pars]})
  fields=[];calls=[]
  for p in pars:
   t=typ(p['schema'],n+name(p['name'])); f=name(p['name'])
   if p.get('required',False):fields.append('\t'+f+' '+t);calls.append('p["'+p['name']+'"] = req.'+f)
   else:fields.append('\t'+f+' Optional['+t+']');calls.append('if req.'+f+'.Set { p["'+p['name']+'"] = req.'+f+'.Value }')
  if bodytype:fields.append('\tBody '+bodytype)
  wrappers.append('type '+n+'Params struct {\n'+'\n'.join(fields)+'\n}\n')
  body='req.Body' if bodytype else 'nil'
  start='func (c *Client) '+n+'(ctx context.Context, req '+n+'Params) '
  if kind=='sse':code=start+'(*EventStream, error) { return c.Events(ctx, req.ChatID, StreamOptions{After:req.After.Value}) }\n'
  else:
   ret='error' if not rt else '('+rt+', error)'
   code=start+ret+' {\n p := map[string]any{}\n'+'\n'.join(calls)+'\n'
   if kind=='binary':code+='return c.download(ctx, "'+oid+'", p, '+body+')\n}\n'
   elif not rt:code+='return c.call(ctx, "'+oid+'", p, '+body+', nil)\n}\n'
   else:code+='var out '+rt+'\nerr := c.call(ctx, "'+oid+'", p, '+body+', &out)\nreturn out, err\n}\n'
  wrappers.append(code)
while extra:
 n,s=extra.popitem()
 if n in [name(x) for x in schemas]:continue
 models.append(decl(n,s))
(root/'models_generated.go').write_text('package envai\n\nimport "encoding/json"\n\n'+'\n'.join(models))
(root/'operations_generated.go').write_text('package envai\n\nimport ("context";"encoding/json")\n\n'+'\n'.join(wrappers))
(root/'contract/operations.json').write_text(json.dumps(ops,indent=2)+'\n')
print(len(schemas),'models;',len(ops),'operations')
def sample(s,depth=0):
 if depth>15:return None
 if '$ref' in s:return sample(schemas[s['$ref'].rsplit('/',1)[-1]],depth+1)
 if 'example' in s:return s['example']
 if 'const' in s:return s['const']
 if 'enum' in s:return s['enum'][0]
 if 'anyOf' in s and not s.get('properties'):return sample(next((x for x in s['anyOf'] if x.get('type')!='null'),{}),depth+1)
 if 'oneOf' in s and not s.get('properties'):return sample(s['oneOf'][0],depth+1)
 t=s.get('type');t=next((x for x in t if x!='null'),None) if isinstance(t,list) else t
 if t=='array':return []
 if t=='object' or 'properties' in s:return {k:sample(v,depth+1) for k,v in s.get('properties',{}).items()}
 if t=='string':return 'fixture'
 if t in ['integer','number']:return max(s.get('minimum',1),1)
 if t=='boolean':return True
 return {}
cases=[]
for path,methods in api['paths'].items():
 for method,op in methods.items():
  if not isinstance(op,dict) or 'operationId' not in op:continue
  oid=op['operationId'];entry={'id':oid,'go_name':name(oid),'parameters':{},'body':None,'status':204,'response':None}
  for pp in op.get('parameters',[]):
   p=api['components']['parameters'][pp['$ref'].rsplit('/',1)[-1]] if '$ref' in pp else pp
   entry['parameters'][name(p['name'])]=sample(p['schema'])
  content=op.get('requestBody',{}).get('content',{})
  if 'application/json' in content:entry['body']=sample(content['application/json']['schema'])
  for status,r in op.get('responses',{}).items():
   if status.startswith('2'):
    entry['status']=int(status)
    if 'application/json' in r.get('content',{}):entry['response']=sample(r['content']['application/json']['schema'])
    break
  if oid=='resolveChatApproval':entry['status']=409;entry['response']={'error':{'type':'conflict','message':'approval unavailable'}}
  cases.append(entry)
(root/'testdata/api-cases.json').write_text(json.dumps(cases,indent=2)+'\n')

ev=json.loads((root/'contract/events.json').read_text())['$defs']
eventmodels=[];switch=[]
for key,schema in ev.items():
 if key in ['Event','Envelope']:continue
 n='ATP'+name(key)
 eventmodels.append(decl(n,schema))
 if '.' in key or key=='ping':switch.append('case "'+key+'":var value '+n+';err:=json.Unmarshal(e.Data,&value);return value,err')
while extra:
 n,s=extra.popitem();eventmodels.append(decl(n,s))
(root/'events_generated.go').write_text('package envai\n\nimport "encoding/json"\n\n'+'\n'.join(eventmodels)+'\nfunc(e Event) TypedPayload()(any,error){switch e.Type {\n'+'\n'.join(switch)+'\n};return append(json.RawMessage(nil),e.Data...),nil}\n')
