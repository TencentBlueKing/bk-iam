#!/bin/bash

echo "generate test environment_test.json"

sed 's#127.0.0.1:9000#127.0.0.1:9200#g' iam.postman_environment.json > iam.postman_environment_test.json


echo "run api test"
newman run -e iam.postman_environment_test.json iam.postman_collection.json
RESULT=$?

echo "the execute result: $RESULT"

exit $RESULT
