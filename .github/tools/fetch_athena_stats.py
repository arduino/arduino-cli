import boto3
import semver
import os
import logging
import uuid
import time


# logging.basicConfig(stream=sys.stdout, level=logging.DEBUG)
log = logging.getLogger()
logging.getLogger("boto3").setLevel(logging.CRITICAL)
logging.getLogger("botocore").setLevel(logging.CRITICAL)
logging.getLogger("urllib3").setLevel(logging.CRITICAL)


def execute(client, statement, dest_s3_output_location):
    log.info("execute query: {} dumping in {}".format(statement, dest_s3_output_location))
    result = client.start_query_execution(
        QueryString=statement,
        ClientRequestToken=str(uuid.uuid4()),
        QueryExecutionContext={"Database": "etl_kpi_prod_hwfw"},
        ResultConfiguration={
            "OutputLocation": dest_s3_output_location,
        },
    )
    execution_id = result["QueryExecutionId"]
    log.info("wait for query {} completion".format(execution_id))
    wait_for_query_execution_completion(client, execution_id)
    log.info("operation successful")
    return execution_id


def wait_for_query_execution_completion(client, query_execution_id):
    query_ended = False
    while not query_ended:
        query_execution = client.get_query_execution(QueryExecutionId=query_execution_id)
        state = query_execution["QueryExecution"]["Status"]["State"]
        if state == "SUCCEEDED":
            query_ended = True
        elif state in ["FAILED", "CANCELLED"]:
            raise BaseException(
                "query failed or canceled: {}".format(query_execution["QueryExecution"]["Status"]["StateChangeReason"])
            )
        else:
            time.sleep(1)


def valid(key):
    split = key.split("_")
    if len(split) < 1:
        return False
    try:
        semver.parse(split[0])
    except ValueError:
        return False
    return True


def get_results(client, execution_id):
    results_paginator = client.get_paginator("get_query_results")
    results_iter = results_paginator.paginate(QueryExecutionId=execution_id, PaginationConfig={"PageSize": 1000})
    res = {}
    for results_page in results_iter:
        for row in results_page["ResultSet"]["Rows"][1:]:
            # Loop through the JSON objects
            key = row["Data"][0]["VarCharValue"]
            if valid(key):
                res[key] = row["Data"][1]["VarCharValue"]

    return res


def convert_data(data):
    result = []
    for key, value in data.items():
        # 0.18.0_macOS_64bit.tar.gz
        split_key = key.split("_")
        if len(split_key) != 3:
            continue
        (version, os_version, arch) = split_key
        arch_split = arch.split(".")
        if len(arch_split) < 1:
            continue
        arch = arch_split[0]
        if len(arch) > 10:
            # This can't be an architecture really.
            # It's an ugly solution but works for now so deal with it.
            continue
        repo = os.environ["GITHUB_REPOSITORY"].split("/")[1]
        result.append(
            {
                "type": "gauge",
                "name": "arduino.downloads.total",
                "value": value,
                "host": os.environ["GITHUB_REPOSITORY"],
                "tags": [
                    f"version:{version}",
                    f"os:{os_version}",
                    f"arch:{arch}",
                    "cdn:downloads.arduino.cc",
                    f"project:{repo}",
                ],
            }
        )

    return result


if __name__ == "__main__":
    DEST_S3_OUTPUT = os.environ["AWS_ATHENA_OUTPUT_LOCATION"]
    AWS_ATHENA_SOURCE_TABLE = os.environ["AWS_ATHENA_SOURCE_TABLE"]

    session = boto3.session.Session(region_name="us-east-1")
    athena_client = session.client("athena")

    query = f"""SELECT replace(json_extract_scalar(url_decode(url_decode(querystring)),
'$.data.url'), 'https://downloads.arduino.cc/arduino-cli/arduino-cli_', '')
AS flavor, count(json_extract(url_decode(url_decode(querystring)),'$')) AS gauge
FROM {AWS_ATHENA_SOURCE_TABLE}
WHERE json_extract_scalar(url_decode(url_decode(querystring)),'$.data.url')
LIKE 'https://downloads.arduino.cc/arduino-cli/arduino-cli_%'
AND json_extract_scalar(url_decode(url_decode(querystring)),'$.data.url')
NOT LIKE '%latest%' -- exclude latest redirect
AND json_extract_scalar(url_decode(url_decode(querystring)),'$.data.url')
NOT LIKE '%alpha%' -- exclude early alpha releases
AND json_extract_scalar(url_decode(url_decode(querystring)),'$.data.url')
NOT LIKE '%.tar.bz2%' -- exclude very old releases archive formats
group by 1 ;"""
    exec_id = execute(athena_client, query, DEST_S3_OUTPUT)
    results = get_results(athena_client, exec_id)
    result_json = convert_data(results)

    print(f"::set-output name=result::{result_json}")
