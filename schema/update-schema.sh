#!/usr/bin/env bash
#
# Downloads the latest OPC UA schema files from the OPC Foundation GitHub repository
# into the schema directory. Run manually when the spec is updated:
#
#   ./schema/update-schema.sh
#
# After downloading, regenerate code with: make gen

set -o errexit
set -o nounset
set -o pipefail

script_dir=$(cd "$(dirname "$0")" && pwd)

echo "Downloading NodeIds.csv"
wget -nv https://raw.githubusercontent.com/OPCFoundation/UA-Nodeset/master/Schema/NodeIds.csv -O "${script_dir}/NodeIds.csv"

echo "Downloading StatusCode.csv"
wget -nv https://raw.githubusercontent.com/OPCFoundation/UA-Nodeset/master/Schema/StatusCode.csv -O "${script_dir}/StatusCode.csv"

echo "Downloading AttributeIds.csv"
wget -nv https://raw.githubusercontent.com/OPCFoundation/UA-Nodeset/master/Schema/AttributeIds.csv -O "${script_dir}/AttributeIds.csv"

echo "Downloading ServerCapabilities.csv"
wget -nv https://raw.githubusercontent.com/OPCFoundation/UA-Nodeset/master/Schema/ServerCapabilities.csv -O "${script_dir}/ServerCapabilities.csv"

echo "Downloading Opc.Ua.NodeIds.permissions.csv"
wget -nv https://raw.githubusercontent.com/OPCFoundation/UA-Nodeset/master/Schema/Opc.Ua.NodeIds.permissions.csv -O "${script_dir}/Opc.Ua.NodeIds.permissions.csv"

echo "Downloading Opc.Ua.Types.bsd"
wget -nv https://raw.githubusercontent.com/OPCFoundation/UA-Nodeset/master/Schema/Opc.Ua.Types.bsd -O "${script_dir}/Opc.Ua.Types.bsd"

echo "Downloading Opc.Ua.NodeSet2.xml"
wget -nv https://raw.githubusercontent.com/OPCFoundation/UA-Nodeset/master/Schema/Opc.Ua.NodeSet2.xml -O "${script_dir}/Opc.Ua.NodeSet2.xml"

echo "Downloading Opc.Ua.PredefinedNodes.xml"
wget -nv https://raw.githubusercontent.com/OPCFoundation/UA-Nodeset/master/DotNet/Opc.Ua.PredefinedNodes.xml -O "${script_dir}/Opc.Ua.PredefinedNodes.xml"
