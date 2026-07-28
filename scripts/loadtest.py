#!/usr/bin/env python3
"""Async load-test for the oteldemo services.

Sends a configurable mix of GET /users/{id}/dashboard and
POST /users/{id}/transactions requests to service1 and reports status
distribution + latency metrics.

Usage:
    python3 scripts/loadtest.py --host localhost:8080 --total 1000 --concurrency 50 --mix 50:50
    python3 scripts/loadtest.py --json            # machine-readable output
    python3 scripts/loadtest.py --mix 0:100       # creates only
"""
import argparse
import asyncio
import json
import random
import time
from collections import Counter

try:
    import aiohttp
except ImportError:
    raise SystemExit("aiohttp is required:  pip install aiohttp")


MERCHANTS = [
    "Acme Payroll", "StreamFlix", "Coffee Corner", "PowerGrid Co",
    "Tuneify", "Wire Transfer LTD", "Online Shop", "Book Store",
]
DESCRIPTIONS = [
    "Monthly salary", "Subscription", "Groceries", "Coffee",
    "Electricity bill", "One-time purchase", "Refund", "Transfer",
]
CURRENCIES = ["USD"]
TX_TYPES = ["debit", "credit"]
USER_IDS = [1, 2, 3]


def build_get(base_url):
    uid = random.choice(USER_IDS)
    return ("GET", f"{base_url}/api/v1/users/{uid}/dashboard", None)


def build_post(base_url):
    uid = random.choice(USER_IDS)
    payload = {
        "amount": round(random.uniform(1, 1000), 2),
        "currency": random.choice(CURRENCIES),
        "type": random.choice(TX_TYPES),
        "merchant": random.choice(MERCHANTS),
        "description": random.choice(DESCRIPTIONS),
    }
    return ("POST", f"{base_url}/api/v1/users/{uid}/transactions", payload)


def percentile(sorted_vals, p):
    if not sorted_vals:
        return 0.0
    if len(sorted_vals) == 1:
        return sorted_vals[0]
    k = (len(sorted_vals) - 1) * (p / 100.0)
    f = int(k)
    c = min(f + 1, len(sorted_vals) - 1)
    if f == c:
        return sorted_vals[f]
    return sorted_vals[f] + (sorted_vals[c] - sorted_vals[f]) * (k - f)


def latencies(results, endpoint=None):
    return [r["latency_ms"] for r in results if endpoint is None or r["endpoint"] == endpoint]


async def fire(session, sem, req, results):
    method, url, payload = req
    endpoint = "GET dashboard" if method == "GET" else "POST create"
    async with sem:
        start = time.perf_counter()
        err = None
        try:
            if method == "GET":
                async with session.get(url) as resp:
                    await resp.read()
                    status = resp.status
            else:
                async with session.post(url, json=payload) as resp:
                    await resp.read()
                    status = resp.status
        except Exception as e:  # network error / timeout / connection refused
            status = 0
            err = str(e)
        elapsed_ms = (time.perf_counter() - start) * 1000.0
        results.append({"endpoint": endpoint, "status": status, "latency_ms": elapsed_ms, "error": err})


async def run(args):
    base_url = f"http://{args.host}"

    parts = args.mix.split(":")
    if len(parts) != 2:
        raise SystemExit("--mix must be like 50:50 (get:post)")
    get_pct, post_pct = int(parts[0]), int(parts[1])
    total_pct = get_pct + post_pct
    if total_pct <= 0:
        raise SystemExit("--mix ratio must sum to > 0")
    n_get = round(args.total * get_pct / total_pct)
    n_post = args.total - n_get

    requests = [build_get(base_url) for _ in range(n_get)]
    requests += [build_post(base_url) for _ in range(n_post)]
    random.shuffle(requests)

    sem = asyncio.Semaphore(args.concurrency)
    results = []

    print(f"Sending {len(requests)} requests ({n_get} GET, {n_post} POST) "
          f"to {base_url} with concurrency {args.concurrency} ...")

    connector = aiohttp.TCPConnector(limit=args.concurrency)
    timeout = aiohttp.ClientTimeout(total=30)
    start = time.perf_counter()
    async with aiohttp.ClientSession(connector=connector, timeout=timeout) as session:
        await asyncio.gather(*(fire(session, sem, r, results) for r in requests))
    elapsed = time.perf_counter() - start

    report = build_report(results, elapsed)
    if args.json:
        print(json.dumps(report, indent=2))
    else:
        print_report(report)


def build_report(results, elapsed):
    total = len(results)
    by_status = Counter(r["status"] for r in results)
    by_class = Counter()
    for s, c in by_status.items():
        key = "0xx (error)" if s == 0 else f"{s // 100}xx"
        by_class[key] += c
    errors = sum(1 for r in results if r["error"])
    success = sum(1 for r in results if 200 <= r["status"] < 400)

    all_lat = sorted(latencies(results))
    rps = total / elapsed if elapsed > 0 else 0.0

    report = {
        "total_requests": total,
        "elapsed_s": round(elapsed, 3),
        "throughput_rps": round(rps, 2),
        "success_rate": round(success / total * 100, 2) if total else 0.0,
        "errors": errors,
        "status_codes": dict(sorted(by_status.items(), key=lambda kv: str(kv[0]))),
        "status_classes": dict(sorted(by_class.items())),
        "latency_ms": _lat_block(all_lat),
        "per_endpoint": {},
    }
    for ep in ["GET dashboard", "POST create"]:
        lat = sorted(latencies(results, ep))
        if not lat:
            continue
        ep_statuses = Counter(r["status"] for r in results if r["endpoint"] == ep)
        report["per_endpoint"][ep] = {
            "count": len(lat),
            "status_codes": dict(sorted(ep_statuses.items(), key=lambda kv: str(kv[0]))),
            "latency_ms": _lat_block(lat),
        }
    return report


def _lat_block(sorted_vals):
    if not sorted_vals:
        return {"min": 0, "mean": 0, "p50": 0, "p90": 0, "p95": 0, "p99": 0, "max": 0}
    return {
        "min": round(sorted_vals[0], 3),
        "mean": round(sum(sorted_vals) / len(sorted_vals), 3),
        "p50": round(percentile(sorted_vals, 50), 3),
        "p90": round(percentile(sorted_vals, 90), 3),
        "p95": round(percentile(sorted_vals, 95), 3),
        "p99": round(percentile(sorted_vals, 99), 3),
        "max": round(sorted_vals[-1], 3),
    }


def print_report(r):
    print("\n" + "=" * 55)
    print("LOAD TEST RESULTS")
    print("=" * 55)
    print(f"Total requests : {r['total_requests']}")
    print(f"Elapsed        : {r['elapsed_s']} s")
    print(f"Throughput     : {r['throughput_rps']} req/s")
    print(f"Success rate   : {r['success_rate']}%")
    print(f"Errors         : {r['errors']}")
    print("\nStatus codes:")
    for s, c in r["status_codes"].items():
        print(f"  {s}: {c}")
    print("\nStatus classes:")
    for s, c in r["status_classes"].items():
        print(f"  {s}: {c}")
    print("\nLatency (ms) - all:")
    for k, v in r["latency_ms"].items():
        print(f"  {k:5}: {v}")
    for ep, data in r["per_endpoint"].items():
        print(f"\nPer endpoint: {ep}")
        print(f"  count : {data['count']}")
        print(f"  status: {data['status_codes']}")
        print("  latency (ms):")
        for k, v in data["latency_ms"].items():
            print(f"    {k:5}: {v}")


def main():
    p = argparse.ArgumentParser(description="Async load-test for oteldemo services")
    p.add_argument("--host", default="localhost:8080", help="service1 host:port")
    p.add_argument("--total", type=int, default=1000, help="total requests")
    p.add_argument("--concurrency", type=int, default=50, help="max concurrent requests")
    p.add_argument("--mix", default="50:50", help="get:post ratio, e.g. 50:50")
    p.add_argument("--json", action="store_true", help="output JSON instead of a table")
    p.add_argument("--seed", type=int, default=None, help="random seed for reproducibility")
    args = p.parse_args()
    if args.seed is not None:
        random.seed(args.seed)
    asyncio.run(run(args))


if __name__ == "__main__":
    main()
