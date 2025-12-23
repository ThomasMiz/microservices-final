#!/usr/bin/python3

import http.server
import socketserver
import urllib.request
import urllib.error
import sys
import argparse
import functools
import os

import urllib.parse

class ProxyHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, api_base=None, **kwargs):
        self.api_base = api_base.rstrip('/')
        parsed_url = urllib.parse.urlparse(self.api_base)
        self.host_addr = parsed_url.netloc
        super().__init__(*args, **kwargs)

    def do_API_proxy(self):
        # Strip /api prefix from local path and append to api_base
        target_path = self.path
        if target_path.startswith('/api'):
            target_path = target_path[4:]
            
        url = self.api_base + target_path
        
        # Prepare headers
        headers = {}
        for key, val in self.headers.items():
            # Filter out Accept-Encoding to prevent upstream from sending compressed data (gzip/br)
            # as we are stripping Content-Encoding headers in the response
            if key.lower() == 'accept-encoding':
                continue
            headers[key] = val

        # Only overwrite Host if necessary, usually safe to match target or leave it
        headers['Host'] = self.host_addr
        
        content_len = int(self.headers.get('Content-Length', 0))
        post_body = self.rfile.read(content_len) if content_len > 0 else None
        
        try:
            req = urllib.request.Request(url, data=post_body, headers=headers, method=self.command)
            with urllib.request.urlopen(req) as response:
                self.send_response(response.status)
                for key, val in response.headers.items():
                    if key.lower() not in ['transfer-encoding', 'content-encoding', 'content-length', 'access-control-allow-origin']:
                        self.send_header(key, val)
                
                # Add CORS headers just in case we need them, though same-origin doesn't strict require them
                self.send_header('Access-Control-Allow-Origin', '*')
                self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS, PUT, DELETE')
                self.send_header('Access-Control-Allow-Headers', '*')
                
                # Calculate length to be sure
                body = response.read()
                self.send_header('Content-Length', len(body))
                self.end_headers()
                self.wfile.write(body)
                
        except urllib.error.HTTPError as e:
            self.send_response(e.code)
            self.end_headers()
            self.wfile.write(e.read())
        except Exception as e:
            self.send_response(500)
            self.end_headers()
            self.wfile.write(str(e).encode())

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS, PUT, DELETE')
        self.send_header('Access-Control-Allow-Headers', '*')
        self.end_headers()

    def do_GET(self):
        if self.path.startswith('/api'):
            self.do_API_proxy()
        else:
            super().do_GET()

    def do_POST(self):
        if self.path.startswith('/api'):
            self.do_API_proxy()
        else:
            self.send_error(405)

    def do_PUT(self):
        if self.path.startswith('/api'):
            self.do_API_proxy()
        else:
            self.send_error(405)

    def do_DELETE(self):
        if self.path.startswith('/api'):
            self.do_API_proxy()
        else:
            self.send_error(405)


def parse_arguments():
    DEFAULTS = {
        'api_base': 'http://34.207.169.29:30526',
        'local_port': 24682
    }

    parser = argparse.ArgumentParser(description='Run C-Hotel Frontend via proxy server')
    parser.add_argument('--api-base', type=str, required=False, default=DEFAULTS['api_base'], help='The target base URL to forward requests to')
    parser.add_argument('-p', '--local-port', type=int, required=False, default=DEFAULTS['local_port'], help='The port to listen on')
    return parser.parse_args()


def create_config_js() -> None:
    path = os.path.join(os.path.dirname(__file__), 'js', 'config.js')
    with open(path, 'w') as f:
        f.write('// THIS FILE IS AUTO-GENERATED AND MUST NOT BE EDITED MANUALLY\n\n')
        f.write(f'const DEFAULT_API_BASE = "/api";\n')


def serve_frontend(local_port: int, api_base: str) -> None:
    handler = functools.partial(ProxyHTTPRequestHandler, api_base=api_base)
    with socketserver.TCPServer(("", local_port), handler) as httpd:
        httpd.serve_forever()


if __name__ == '__main__':
    try:
        args = parse_arguments()

        print(f'Starting C-Hotel Frontend using proxy server')
        print(f' - Serving on port {args.local_port}')
        print(f' - Proxying /api to {args.api_base}')
        print()

        create_config_js()
        serve_frontend(args.local_port, args.api_base)
    except KeyboardInterrupt:
        print('\nShutting down server...')
