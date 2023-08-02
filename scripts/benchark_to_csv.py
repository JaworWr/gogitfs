import argparse
import csv
import re
import sys
from pathlib import Path
from typing import Iterable


BENCHMARK_RE = re.compile(
    re.escape("[BENCHMARK]") + r" (.+): ([0-9.]+.{1,2}) \(([0-9.]+)ms\)$"
)


def get_parser():
    parser = argparse.ArgumentParser()
    parser.add_argument("-f", "--funcs", nargs="+", help="functions to extract. By default - all.")
    parser.add_argument("logfiles", nargs="+", help="log files to process", type=Path)
    return parser


def get_times(lines: Iterable[str], funcs):
    for line in lines:
        m = BENCHMARK_RE.search(line)
        if not m:
            continue
        func, htime, time = m.groups()
        if funcs is not None and func not in funcs:
            continue
        yield func, htime, float(time)


def times_to_csv(file, times: Iterable[tuple[str, str, float]]):
    fieldnames = ["function", "time", "time_ms"]
    writer = csv.DictWriter(file, fieldnames)
    writer.writeheader()
    writer.writerows(dict(zip(fieldnames, t)) for t in times)


def process_logfile(path: Path, funcs) -> Path:
    path_out = path.with_suffix(".csv")
    with open(path) as f_in, open(path_out, "w") as f_out:
        times = get_times(f_in, funcs)
        times_to_csv(f_out, times)
    return path_out


def main(args):
    for path in args.logfiles:
        out_path = process_logfile(path, args.funcs)
        print("Processed", path, "->", out_path, file=sys.stderr, flush=True)


if __name__ == "__main__":
    args = get_parser().parse_args()
    main(args)
