"""Verify exact shared schemas, attributed fixtures, and generated operation inventory."""
import hashlib
import json
import pathlib
import re

# Exact reviewed backend source pin; this does not imply a published release.
EXPECTED_CONTRACT_REPOSITORY = 'digitalygo/envai-teams'
EXPECTED_CONTRACT_COMMIT = '7f4085dbfdd3f208b82716119cff1860fbd502bb'
REQUIRED_OPERATIONS = {
    'deleteWorkspace', 'updateChat', 'deleteChat', 'listChatCrons', 'getChatCron',
    'updateChatCron', 'deleteChatCron', 'listProjectSkills', 'createProjectSkill',
    'getProjectSkill', 'updateProjectSkill', 'deleteProjectSkill',
    'getProjectUsage', 'getProjectLogs',
}
EXCLUDED_HUMAN_ADMIN_OPERATIONS = {
    'getAuthCsrf', 'registerUser', 'registerFromInvitation', 'login', 'forgotPassword',
    'resetPassword', 'logout', 'getAuthenticatedUser', 'sendEmailVerification',
    'verifyEmail', 'resolveInvitation', 'acceptInvitation', 'listCompanies',
    'createCompany', 'getCompany', 'updateCompany', 'deleteCompany', 'getCompanyUsage',
    'listCompanyMembers', 'updateCompanyMember', 'deleteCompanyMember',
    'listCompanyInvitations', 'inviteCompanyMember', 'deleteCompanyInvitation',
    'resendCompanyInvitation', 'listCompanyProjects', 'createProject',
    'getCompanyBillingWallet', 'listCompanyBillingTransactions',
    'quoteCompanyBillingTopup', 'createCompanyBillingTopup',
    'setupCompanyBillingPaymentMethod', 'updateCompanyAutoRecharge',
    'receiveStripeBillingWebhook', 'getProjectSpendLimit', 'updateProjectSpendLimit',
    'getApiKeySpendLimit', 'updateApiKeySpendLimit', 'getProject', 'updateProject',
    'deleteProject', 'listProjectMembers', 'assignProjectMember',
    'unassignProjectMember', 'listUsers', 'getUser', 'updateUserPlatformRole',
    'receiveElevenLabsCallback', 'internalWikiDispatch',
    'internalChatAttachmentResolve', 'checkAdmissionBilling', 'reserveBilling',
    'settleBilling', 'releaseBilling', 'createChatCron',
}
API_KEY_STUBS = {'createApiKey', 'listApiKeys', 'deleteApiKey'}

def verify(root):
    provenance = json.loads((root / 'contract/provenance.json').read_text())
    expected_paths = {name: 'packages/contracts/' + name for name in (
        'openapi.yaml', 'events.json', 'events-v2.json', 'wiki.json', 'terminal-snapshot.json', 'smidja-native.json')}
    assert provenance['repository'] == EXPECTED_CONTRACT_REPOSITORY
    assert provenance['commit'] == EXPECTED_CONTRACT_COMMIT
    if EXPECTED_CONTRACT_COMMIT is None:
        assert provenance['status'] == 'working-tree snapshot awaiting guarded backend publication'
    else:
        assert re.fullmatch('[0-9a-f]{40}', EXPECTED_CONTRACT_COMMIT)
    assert provenance['protocol_version'] == 3 and provenance['operations'] == 92
    assert provenance['server_operations'] == 146
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
    assert sources['openapi.yaml']['git_blob_sha'] == 'fa965e896047f94084ddc7160a5a3cbf9633b986'
    # Original numeric v2 is immutable, independently of the new source pin.
    assert sources['events-v2.json']['git_blob_sha'] == '02b13b23cd06eacd81059d1ed99e0027769d75e4'
    fixtures = provenance['generated_fixtures']
    assert set(fixtures) == {'events-invalid.json', 'events-valid.json', 'operations.json', 'signatures.json',
                            'sse-cases.json', 'api-cases.json', 'operation-details.json', 'schema-matrix.json', 'openapi.json'}
    shared = provenance['shared_fixtures']
    expected_shared = {str(p.relative_to(root / 'testdata')) for directory in ('wiki', 'completion-terminal', 'control-plane-http', 'v3-http', 'chat-attachments', 'parity')
                       for p in (root / 'testdata' / directory).glob('*.json')}
    assert set(shared) == expected_shared, 'Incomplete shared fixture inventory'
    for collection in (fixtures, shared):
        for name, entry in collection.items():
            assert entry['origin'], name + ' origin'
            data = (root / entry['snapshot_path']).read_bytes()
            assert hashlib.sha256(data).hexdigest() == entry['sha256'], name + ' SHA256'
    legacy = provenance['legacy_artifacts']['completion-v2.json']
    assert legacy['source_commit'] == '55275860c9a950eaad8e39e9d2b728149084779c'
    assert legacy['source_git_blob_sha'] == '2d3fe7581229a6fb7a798c99ecccfe5f823fa6ed'
    assert hashlib.sha256((root / legacy['snapshot_path']).read_bytes()).hexdigest() == legacy['sha256']
    actual = json.loads((root / 'contract/operations.json').read_text())
    expected = json.loads((root / 'testdata/operations.json').read_text())
    allowed = set(json.loads((root / 'scripts/sdk-operations.json').read_text()))
    ids = {op['id'] for op in actual}
    assert len(allowed) == 92 and len(actual) == 92 and len(expected) == 92
    assert ids == allowed and REQUIRED_OPERATIONS <= ids
    assert API_KEY_STUBS <= ids
    assert EXCLUDED_HUMAN_ADMIN_OPERATIONS.isdisjoint(ids)
    assert 'getProjectLogs' in ids
    assert [{k: op[k] for k in ['id', 'method', 'path', 'authenticated']} for op in actual] == expected
    logs = next(op for op in actual if op['id'] == 'getProjectLogs')
    assert {p['name'] for p in logs['parameters'] if p['repeated']} == {'types[]', 'workspace_ids[]', 'api_key_ids[]'}

if __name__ == '__main__':
    verify(pathlib.Path(__file__).resolve().parent.parent)
    print('All 6 shared contracts, attributed fixtures and 92-operation inventory verified')
