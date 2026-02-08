#!/bin/bash
# Mock Sidecar Plugin
# Reads JSON from stdin, writes JSON to stdout

while read line; do
  # Simple echo logic: return data with added "echo" field
  # Using jq if available, else python or manual
  # Since this is minimal env, let's use python3
  python3 -c "
import sys, json
line = sys.stdin.readline()
if not line: exit(0)
msg = json.loads(line)
if msg['type'] == 'guest_call':
    resp = {
        'type': 'guest_response',
        'id': msg['id'],
        'success': True,
        'data': {'echo': msg['parameters'], 'message': 'Hello from Sidecar'},
        'error': None
    }
    print(json.dumps(resp))
    sys.stdout.flush()
" <<< "$line"
done
