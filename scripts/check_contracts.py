"""Verify exact shared schemas, attributed fixtures, and generated operation inventory."""
import hashlib
import json
import pathlib
import re

# Replaced with the verified backend publication commit by the release coordinator.
EXPECTED_CONTRACT_COMMIT = '7192fc05ebec817bddc705186c8f195cc0dc5ec5'

def verify(root):
    provenance = json.loads((root / 'contract/provenance.json').read_text())
    expected_paths = {name: 'packages/contracts/' + name for name in (
        'openapi.yaml', 'events.json', 'events-v2.json', 'wiki.json', 'terminal-snapshot.json')}
    assert provenance['repository'] == 'digitalygo/envai-teams'
    assert provenance['commit'] == EXPECTED_CONTRACT_COMMIT
    if EXPECTED_CONTRACT_COMMIT is None:
        assert provenance['status'] == 'working-tree snapshot awaiting guarded backend publication'
    else:
        assert re.fullmatch('[0-9a-f]{40}', EXPECTED_CONTRACT_COMMIT)
    assert provenance['protocol_version'] == 3 and provenance['operations'] == 71
    sources = provenance['sources']
    assert set(sources) == set(expected_paths), 'Incomplete source inventory'
    assert 'files' not in provenance, 'Sources and fixtures must be separated'
    for name, entry in sources.items():
        assert entry['repository_path'] == expected_paths[name], name
        assert entry['snapshot_path'] == 'contract/' + name, name
        data = (root / entry['snapshot_path']).read_bytes()
        assert len(data) == entry['bytes'], name + ' byte count'
        assert hashlib.sha256(data).hexdigest() == entry['sha256'], name + ' SHA256'
        assert hashlib.sha1(b'blob ' + str(len(data)).encode() + b'\0' + data).hexdigest() == entry['git_blob_sha'], name + ' Git blob'
    # Original numeric v2 is immutable, independently of the new source pin.
    assert sources['events-v2.json']['git_blob_sha'] == '02b13b23cd06eacd81059d1ed99e0027769d75e4'
    fixtures = provenance['generated_fixtures']
    assert set(fixtures) == {'events-invalid.json', 'events-valid.json', 'operations.json', 'signatures.json',
                            'sse-cases.json', 'api-cases.json', 'operation-details.json', 'schema-matrix.json', 'openapi.json'}
    shared = provenance['shared_fixtures']
    expected_shared = {str(p.relative_to(root / 'testdata')) for directory in ('wiki', 'completion-terminal', 'control-plane-http', 'v3-http', 'chat-attachments')
                       for p in (root / 'testdata' / directory).glob('*.json')}
    assert set(shared) == expected_shared, 'Incomplete shared fixture inventory'
    for collection in (fixtures, shared):
        for name, entry in collection.items():
            assert entry['origin'], name + ' origin'
            data = (root / entry['snapshot_path']).read_bytes()
            assert hashlib.sha256(data).hexdigest() == entry['sha256'], name + ' SHA256'
    actual = json.loads((root / 'contract/operations.json').read_text())
    expected = json.loads((root / 'testdata/operations.json').read_text())
    assert len(actual) == 71
    assert [{k: op[k] for k in ['id', 'method', 'path', 'authenticated']} for op in actual] == expected

if __name__ == '__main__':
    verify(pathlib.Path(__file__).resolve().parent.parent)
    print('All 5 shared schemas, attributed fixtures and 71-operation inventory verified')
