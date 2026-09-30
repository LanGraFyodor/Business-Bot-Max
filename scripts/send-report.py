#!/usr/bin/env python3
"""Send an exported business report to Docspace or read its approval status.
Python 3.9+, standard library only. Token: DOCSPACE_INTEGRATION_TOKEN.
No vendor credentials are required: the caller supplies an exported file.
"""
import argparse
import json
import os
from pathlib import Path
import sys
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode, urlsplit
from urllib.request import Request, build_opener, HTTPRedirectHandler

class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        # Never forward an integration credential to a redirected host.
        return None

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--base-url', required=True, help='https://docspace.company.ru')
    action = parser.add_mutually_exclusive_group(required=True)
    action.add_argument('--file', type=Path, help='PDF, XLSX, CSV, DOCX, TXT, PNG or JPG')
    action.add_argument('--status', metavar='EXTERNAL_ID', help='Read the report and its approval status')
    action.add_argument('--channels', action='store_true', help='List channels available to this key')
    parser.add_argument('--channel')
    parser.add_argument('--title')
    parser.add_argument('--external-id', help='Stable unique ID for this report revision')
    parser.add_argument('--review', action='store_true')
    args = parser.parse_args()
    base = args.base_url.rstrip('/')
    parts = urlsplit(base)
    if parts.username or parts.password or parts.query or parts.fragment or parts.path not in ('', '/'):
        parser.error('base-url must contain only scheme, host and optional port')
    if not parts.hostname or not (parts.scheme == 'https' or parts.scheme == 'http' and parts.hostname in ('localhost', '127.0.0.1', '::1')):
        parser.error('HTTPS is required; HTTP is allowed only on loopback for local testing')
    token = os.environ.get('DOCSPACE_INTEGRATION_TOKEN', '').strip()
    if not token:
        parser.error('Set DOCSPACE_INTEGRATION_TOKEN before running this script')
    data = None
    if args.channels:
        path = '/api/integrations/channels'
    elif args.status:
        path = '/api/integrations/documents?' + urlencode({'externalId': args.status})
    else:
        if not args.channel or not args.title or not args.external_id:
            parser.error('--file requires --channel, --title and --external-id')
        if args.file.suffix.lower() not in ('.pdf','.xlsx','.csv','.docx','.txt','.png','.jpg','.jpeg'):
            parser.error('Unsupported file extension; save XLS/MXL as XLSX, PDF or CSV first')
        if not args.file.is_file() or not 0 < args.file.stat().st_size <= 25 * 1024 * 1024:
            parser.error('File must exist and have a size of 1 byte to 25 MB')
        data = args.file.read_bytes()
        path = '/api/integrations/documents?' + urlencode({'channelId':args.channel, 'title':args.title, 'filename':args.file.name, 'externalId':args.external_id, 'review':'1' if args.review else '0'})
    request = Request(base + path, data=data, headers={'Authorization':'Bearer '+token, 'Content-Type':'application/octet-stream'}, method='POST' if data is not None else 'GET')
    try:
        with build_opener(NoRedirect()).open(request, timeout=60) as response:
            result = json.loads(response.read())
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0
    except HTTPError as error:
        print('HTTP %s: %s' % (error.code, error.read(4096).decode('utf-8', errors='replace')), file=sys.stderr)
    except (URLError, OSError, ValueError) as error:
        print('Request failed: %s' % error, file=sys.stderr)
    return 1

if __name__ == '__main__':
    raise SystemExit(main())
