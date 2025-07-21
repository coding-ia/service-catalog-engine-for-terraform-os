import glob
import os
import json
import unittest
from terraform_runner import override_manager


class TestOverrideManager(unittest.TestCase):
    TMP_WORKSPACE_DIR = '/tmp'
    OVERRIDE_FILES_PATTERN = '*.tf.json'

    def test_write_backend_override_happy_path(self):
        # arrange
        provisioned_product_descriptor = 'account-id/pp-id'
        state_bucket = 'state-bucket'
        state_region = 'us-west-2'
        expected_backend_override = {
            'terraform': {
                'backend': {
                    's3': {
                        'bucket': f'{state_bucket}',
                        'key': f'{provisioned_product_descriptor}',
                        'region': f'{state_region}',
                        'profile': 'default'
                    }
                }
            }
        }

        # act
        override_manager.write_backend_override(self.TMP_WORKSPACE_DIR,
                                                provisioned_product_descriptor, state_bucket, state_region)
        with open(f'{self.TMP_WORKSPACE_DIR}/{override_manager.BACKEND_FILE_NAME}', 'r') as json_file:
            actual_backend_override = json.load(json_file)

        # assert
        self.assertEqual(expected_backend_override, actual_backend_override)

    def test_write_variable_override_happy_path(self):
        # arrange
        variables = [
            {'key': 'key1', 'value': 'value1'},
            {'key': 'key2', 'value': 'value2'}
        ]
        expected_variable_override = {
            'variable': {
                'key1': {'default': 'value1'},
                'key2': {'default': 'value2'}
            }
        }

        # act
        override_manager.write_variable_override(self.TMP_WORKSPACE_DIR, variables)
        with open(f'{self.TMP_WORKSPACE_DIR}/{override_manager.VARIABLE_FILE_NAME}', 'r') as json_file:
            actual_variable_override = json.load(json_file)

        # assert
        self.assertEqual(expected_variable_override, actual_variable_override)

    def test_write_variable_override_with_complex_type_happy_path(self):
        # arrange
        complex_type_value = {
            'nested_key_1': 'nested_value_1',
            'nested_key_2': 'nested_value_2'
        }
        variables = [
            {'key': 'key1', 'value': 'value1'},
            {'key': 'key2', 'value': json.dumps(complex_type_value)}
        ]
        expected_variable_override = {
            'variable': {
                'key1': {'default': 'value1'},
                'key2': {'default': complex_type_value}
            }
        }

        # act
        override_manager.write_variable_override(self.TMP_WORKSPACE_DIR, variables)
        with open(f'{self.TMP_WORKSPACE_DIR}/{override_manager.VARIABLE_FILE_NAME}', 'r') as json_file:
            actual_variable_override = json.load(json_file)

        # assert
        self.assertEqual(expected_variable_override, actual_variable_override)

    def test_write_variable_override_no_variables(self):
        # arrange
        variables = None

        # act
        override_manager.write_variable_override(self.TMP_WORKSPACE_DIR, variables)

        # assert
        self.assertFalse(
            os.path.exists(f'{self.TMP_WORKSPACE_DIR}/{override_manager.VARIABLE_FILE_NAME}'))

    def test_write_variable_override_empty_variables(self):
        # arrange
        variables = {}

        # act
        override_manager.write_variable_override(self.TMP_WORKSPACE_DIR, variables)

        # assert
        self.assertFalse(
            os.path.exists(f'{self.TMP_WORKSPACE_DIR}/{override_manager.VARIABLE_FILE_NAME}'))

    def tearDown(self):
        # Remove temp files after each test
        override_files = glob.glob(f'{self.TMP_WORKSPACE_DIR}/{self.OVERRIDE_FILES_PATTERN}')
        for override_file in override_files:
            os.remove(override_file)


if __name__ == '__main__':
    unittest.main()
