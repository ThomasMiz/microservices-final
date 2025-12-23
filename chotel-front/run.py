#!/usr/bin/python3

import argparse
import os
import http.server as httpd


def parse_arguments():
    DEFAULTS = {
        'api_base': 'http://tucumano-1010729244.us-east-1.elb.amazonaws.com/api',
        'local_port': 24682
    }

    parser = argparse.ArgumentParser(description='Run C-Hotel Frontend')
    parser.add_argument('--api-base', type=str, required=False, default=DEFAULTS['api_base'], help='API base URL')
    parser.add_argument('-p', '--local-port', type=int, required=False, default=DEFAULTS['local_port'], help='The port to listen on')
    args = parser.parse_args()
    return args


def create_config_js(api_base: str) -> None:
    path = os.path.join(os.path.dirname(__file__), 'js', 'config.js')
    with open(path, 'w') as f:
        f.write('// THIS FILE IS AUTO-GENERATED AND MUST NOT BE EDITED MANUALLY\n\n')
        f.write(f'const DEFAULT_API_BASE = "{api_base}";\n')
        

def serve_frontend(local_port: int) -> None:
    server = httpd.HTTPServer(("", local_port), httpd.SimpleHTTPRequestHandler)
    server.serve_forever()


if __name__ == '__main__':
    try:
        args = parse_arguments()

        print(f'Starting C-Hotel Frontend')
        print(f' - Serving on port {args.local_port}')
        print(f' - API base: {args.api_base}')
        print()

        create_config_js(args.api_base)
        serve_frontend(args.local_port)
    except KeyboardInterrupt:
        print("\nShutting down server...")
