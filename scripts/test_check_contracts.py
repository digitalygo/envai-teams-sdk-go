"""Ensure every provenance entry is checked, including exact trailing bytes."""

import json
import pathlib
import shutil
import tempfile
import unittest

from check_contracts import verify


class ProvenanceTests(unittest.TestCase):
    def test_all_snapshot_bytes_are_checked(self):
        root = pathlib.Path(__file__).resolve().parent.parent
        with tempfile.TemporaryDirectory() as directory:
            copy = pathlib.Path(directory)
            shutil.copytree(root / 'contract', copy / 'contract')
            shutil.copytree(root / 'testdata', copy / 'testdata')
            verify(copy)
            provenance = json.loads((copy / 'contract/provenance.json').read_text())
            for section in ('sources', 'generated_fixtures', 'shared_fixtures'):
                for name, entry in provenance[section].items():
                    with self.subTest(section=section, name=name):
                        path = copy / entry['snapshot_path']
                        original = path.read_bytes()
                        path.write_bytes(original + b'\n')
                        with self.assertRaises(AssertionError):
                            verify(copy)
                        path.write_bytes(original)

    def test_manifest_git_blob_identity_is_checked(self):
        root = pathlib.Path(__file__).resolve().parent.parent
        with tempfile.TemporaryDirectory() as directory:
            copy = pathlib.Path(directory)
            shutil.copytree(root / 'contract', copy / 'contract')
            shutil.copytree(root / 'testdata', copy / 'testdata')
            path = copy / 'contract/provenance.json'
            provenance = json.loads(path.read_text())
            provenance['sources']['openapi.yaml']['git_blob_sha'] = '0' * 40
            path.write_text(json.dumps(provenance))
            with self.assertRaises(AssertionError):
                verify(copy)


if __name__ == '__main__':
    unittest.main()
