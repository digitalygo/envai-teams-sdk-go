"""Verify raw pinned source bytes and separately attributed SDK fixtures."""

import hashlib
import json
import pathlib


def verify(root):
    provenance = json.loads((root / 'contract/provenance.json').read_text())
    sources = provenance['sources']
    expected_paths = {
        'openapi.yaml': 'packages/contracts/openapi.yaml',
        'events.json': 'packages/contracts/events.json',
    }
    assert set(sources) == set(expected_paths), 'Incomplete source inventory'
    assert 'files' not in provenance, 'Source files and generated fixtures must be separated'
    assert provenance['repository'] == 'digitalygo/envai-teams'
    assert provenance['commit'] == '55275860c9a950eaad8e39e9d2b728149084779c'
    assert provenance['protocol_version'] == 2 and provenance['operations'] == 52
    for name, entry in sources.items():
        assert entry['repository_path'] == expected_paths[name], name
        assert entry['snapshot_path'] == 'contract/' + name, name
        data = (root / entry['snapshot_path']).read_bytes()
        assert len(data) == entry['bytes'], name + ' byte count'
        assert hashlib.sha256(data).hexdigest() == entry['sha256'], name + ' SHA256'
        git_header = b'blob ' + str(len(data)).encode() + b'\0'
        assert hashlib.sha1(git_header + data).hexdigest() == entry['git_blob_sha'], name + ' Git blob'
    fixtures = provenance['generated_fixtures']
    assert set(fixtures) == {
        'events-invalid.json', 'events-valid.json', 'operations.json', 'signatures.json',
        'sse-cases.json', 'api-cases.json', 'operation-details.json', 'schema-matrix.json',
    }, 'Incomplete fixture inventory'
    for name, entry in fixtures.items():
        assert entry['origin'], name + ' origin'
        assert 'repository_path' not in entry, name + ' is SDK-generated, not upstream'
        data = (root / entry['snapshot_path']).read_bytes()
        assert hashlib.sha256(data).hexdigest() == entry['sha256'], name + ' SHA256'
    actual = json.loads((root / 'contract/operations.json').read_text())
    expected = json.loads((root / 'testdata/operations.json').read_text())
    assert len(actual) == 52
    assert [{k: op[k] for k in ['id', 'method', 'path', 'authenticated']} for op in actual] == expected


if __name__ == '__main__':
    verify(pathlib.Path(__file__).resolve().parent.parent)
    print('All 2 raw source snapshots, 8 SDK fixtures and 52-operation inventory verified')
