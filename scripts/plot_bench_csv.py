#!/usr/bin/env python3
import argparse
import pandas as pd
import math
import matplotlib
import matplotlib.pyplot as plt
import re
import sys

def parse():
    parser = argparse.ArgumentParser(description='It will plot a csv file created by gobenchchronos')
    parser.add_argument('csv_file', metavar='file', type=str, nargs=1, help='The csv file')
    parser.add_argument('--pkg-regex', metavar='<exp>',dest='pkg_regex', default=None, type=str,
                        help='A regex expresion, only packages that match this expression will be displayed. If empty'
                             ' all packages shown in different subplots')
    parser.add_argument('--bench-regex', dest='bench_regex', metavar='<exp>', default=None, type=str,
                        help='A regexp expression, only benchmarks that match will be shown.')
    parser.add_argument('-max-plots-collums', dest='x_plots', metavar='<int>', default=3, type=int,
                        help='Maximum amount of plots to appear in the x axis')

    return parser.parse_args()


def plot(csv_file: str, pkg_exp=None, bench_exp=None, max_colums=3):
    df = pd.read_csv(csv_file)

    # filter data out based on regex
    pkgs = df['pkg'].unique().tolist()
    if pkg_exp is not None:
        pkgs = [pkg for pkg in pkgs if pkg_exp.search(pkg)]

    df = df[df['pkg'].isin(pkgs)]

    names = df['name'].unique().tolist()
    if bench_exp is not None:
        names = [name for name in names if bench_exp.search(name)]

    df = df[df['name'].isin(names)]
    print(df)

    # each package gets its own subplot so let calculate how many rows
    columns = len(pkgs) if len(pkgs) < max_colums else max_colums
    rows = math.ceil(len(pkgs)/columns)

    for i, pkg in enumerate(pkgs):
        ax = plt.subplot(rows, columns, i+1)
        ax.set_title(f'Benchmarks: {pkg}', y=1.2)
        ax.set_ylabel(f'ns per op')
        ax.set_xlabel(f'commit')
        local_df = df[df['pkg'] == pkg].sort_values(by='ID', ascending=False)
        for name in names:
            name_df = local_df[local_df['name'] == name]
            ax.plot([c[:6] if len(c) > 6 else c for c in name_df['commit'].tolist()], name_df['ns_per_op'].tolist(),
                    label=name)

        ax.grid(True)
        # ax.legend(bbox_to_anchor=(-1, 0), loc='lower left',
        #            ncol=1, fontsize='x-small')


    plt.tight_layout()
    plt.show()


def main():
    args = parse()
    if args.csv_file is None or len(args.csv_file) != 1:
        print('csv_file is required')
        sys.exit(1)

    if args.x_plots < 1:
        print('minimum value for x plots is 1')
        sys.exit(1)

    pkg_regex = None
    if args.pkg_regex is not None:
        try:
            pkg_regex = re.compile(args.pkg_regex)
        except re.error as e:
            print(f'Invalid pkg-filter-regex: {e}')
            sys.exit(1)

    bench_regex = None
    if args.bench_regex is not None:
        try:
            bench_regex = re.compile(args.bench_regex)
        except re.error as e:
            print(f'Invalid bench-regex: {e}')
            sys.exit(1)

    plot(args.csv_file[0], pkg_regex, bench_regex, args.x_plots)


if __name__ == '__main__':
    main()