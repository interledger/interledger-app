# Cleans botanist generatios
find ../typescript/botanist/app/generated/protobuf-ts/ -type f -name '*.client.ts' -exec sh -c 'x="{}"; mv "$x" "${x%.client.ts}_client.ts"' \;

# Cleans backend google generation
find ../go/proto -name 'google' -type d -prune -exec rm -rf '{}' +

