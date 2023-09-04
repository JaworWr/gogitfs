"""Extract benchmark results from logs and save them in CSV files.

For usage see benchmark_to_csv.py -h
"""
import argparse
import csv
import re
import sys
from pathlib import Path
from typing import Iterable

BENCHMARK_RE = re.compile(
    re.escape("[BENCHMARK]") + r" (.+): ([0-9.]+.{1,2}) \(([0-9.]+)ms\)$"
)


def get_parser() -> argparse.ArgumentParser:
    """Get parser for command-line arguments"""
    parser = argparse.ArgumentParser(
        description="Parse application logs and extract running times of benchmarked functions, "
                    "then save them in CSV files."
    )
    parser.add_argument(
        "-f", "--funcs", nargs="+",
        help="names of functions, for which corresponding times should be extracted. By default - all."
    )
    parser.add_argument(
        "logfiles", nargs="+", type=Path,
        help="log files to process. For each file, "
             "a corresponding .csv file will be generated."
    )
    return parser


def get_times(lines: Iterable[str], function_names=None):
    """Extract benchmark times from lines.

    Args:
        lines: lines from log to process
        function_names: names of functions to extract. Pass None to extract times for all functions.

    Returns:
        A generator yielding tuples of function name, human-readable time and time in milliseconds as float.
    """
    for line in lines:
        m = BENCHMARK_RE.search(line)
        if not m:
            continue
        func, htime, time = m.groups()
        if function_names is not None and func not in function_names:
            continue
        yield func, htime, float(time)


def write_times_to_csv(file, times: Iterable[tuple[str, str, float]]):
    fieldnames = ["function", "time", "time_ms"]
    writer = csv.DictWriter(file, fieldnames)
    writer.writeheader()
    writer.writerows(dict(zip(fieldnames, t)) for t in times)


def process_logfile(path: Path, function_names=None) -> Path:
    """Generate CSV file for a single log file.

    Args:
        path: path of the input file
        function_names: names of functions to process - see `get_lines`

    Returns:
        path of the generated CSV file
    """
    path_out = path.with_suffix(".csv")
    with open(path) as f_in, open(path_out, "w") as f_out:
        times = get_times(f_in, function_names)
        write_times_to_csv(f_out, times)
    return path_out


def main(args):
    for path in args.logfiles:
        out_path = process_logfile(path, args.funcs)
        print("Processed", path, "->", out_path, file=sys.stderr, flush=True)


if __name__ == "__main__":
    args = get_parser().parse_args()
    main(args)
