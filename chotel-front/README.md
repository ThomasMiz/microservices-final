# C-Hotel Frontend

## Without Proxy Server

```bash
./run.py [-api_base <api>] [-p <local_port>]
```

Arguments are optional. Default values:
- `api_base`: `http://tucumano-1010729244.us-east-1.elb.amazonaws.com/api`
- `local_port`: `24682`

## With Proxy Server (CORS workaround)

```bash
./run_proxy.py [-api_url <api>] [-p <local_port>]
```

Arguments are optional. Default values:
- `api_url`: `http://34.207.169.29:30526`
- `local_port`: `24682`
