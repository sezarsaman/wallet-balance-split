#!/bin/sh
docker run --rm -v ${PWD}:/app swaggerapi/swagger-codegen-cli-v3 generate \
    -i /app/docs/swagger.yaml \
    -l openapi \
    -o /app/docs/swagger/
mv docs/openapi.json docs/swagger.json
