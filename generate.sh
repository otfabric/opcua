#!/usr/bin/env bash
#
# Regenerates all generated Go source files from the OPC UA schema CSVs and
# type definitions. Run after updating schema files:
#
#   ./generate.sh          # or: make gen
#
# Generators (under cmd/):
#   id          – NodeID constants from NodeIds.csv
#   status      – StatusCode constants from StatusCode.csv
#   attrid      – AttributeID enum from AttributeIds.csv
#   capability  – ServerCapability constants from ServerCapabilities.csv
#   permissions – Default node permissions from Opc.Ua.NodeIds.permissions.csv
#   service     – Service request/response codec from Opc.Ua.Types.bsd
#
# After the generators, stringer is run on all enum types in ua/ and on
# ConnState in the root package.

rm -f ./*/*_gen.go
go run cmd/id/main.go
go run cmd/status/main.go
go run cmd/attrid/main.go
go run cmd/capability/main.go
go run cmd/permissions/main.go
go run cmd/service/*.go

# install stringer if not installed already
command -v stringer || go install golang.org/x/tools/cmd/stringer@latest

# find all enum types
enums=$(grep -w '^type' ua/enums*.go | awk '{print $2;}' | paste -sd, -)

# generate enum string method
(cd ua && stringer -type "$enums" -output enums_strings_gen.go)
echo "Wrote ua/enums_strings_gen.go"

stringer -type ConnState -output connstate_strings_gen.go
echo "Wrote connstate_strings_gen.go"

# remove golang.org/x/tools/cmd/stringer from list of dependencies
go mod tidy
