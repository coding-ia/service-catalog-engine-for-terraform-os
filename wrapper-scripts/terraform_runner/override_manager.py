import json
import os
from json.decoder import JSONDecodeError

BACKEND_FILE_NAME = "backend_override.tf.json"
VARIABLE_FILE_NAME = "variable_override.tf.json"
MAX_SESSION_NAME_LENGTH = 64


def write_backend_override(workspace_dir, provisioned_product_descriptor, state_bucket, state_region):
    backend_override = {
        "terraform": {
            "backend": {
                "s3": {
                    "bucket": f"{state_bucket}",
                    "key": f"{provisioned_product_descriptor}",
                    "region": f"{state_region}",
                    "profile": "default"
                }
            }
        }
    }
    with open(f"{workspace_dir}/{BACKEND_FILE_NAME}", "w") as json_file:
        json.dump(backend_override, json_file)


def write_variable_override(workspace_dir, variables):
    if not variables:
        return

    variable_override = {'variable': {}}
    for variable in variables:
        try:
            variable_value_json = json.loads(variable['value'])
            variable_override['variable'][variable['key']] = {"default": variable_value_json}
        except JSONDecodeError:
            variable_override['variable'][variable['key']] = {"default": variable['value']}
    with open(f"{workspace_dir}/{VARIABLE_FILE_NAME}", "w") as json_file:
        json.dump(variable_override, json_file)


def write_aws_config_file(directory: str, profile: str, region: str, role_arn: str, role_session_name: str) -> str:
    config_template = """[default]
region = {region}
output = json

[profile {profile}]
role_arn = {role_arn}
source_profile = default
role_session_name = {role_session_name}
region = {region}
output = json
"""

    config_content = config_template.format(
        profile=profile,
        region=region,
        role_arn=role_arn,
        role_session_name=__format_session_name(role_session_name)
    )

    config_path = os.path.join(directory, ".config")

    with open(config_path, "w") as f:
        f.write(config_content)

    return config_path

def __format_session_name(unformatted_session_name):
    return f"{unformatted_session_name[:MAX_SESSION_NAME_LENGTH]}".replace('/', '-')
