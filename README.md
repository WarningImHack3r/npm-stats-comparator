# npm-stats-comparator

Compare a few statistics between two GitHub releases of an NPM package. **For now, only the LOCs comparison is available.**  
Includes the intermediate releases between the two specified releases.

_Inspired by [this tweet](https://twitter.com/denlukia/status/1772818790415225202) by [@denlukia](https://github.com/denlukia)._

## Usage

```bash
$ ./npm-stats-comparator --repo user/repo --from v1.0.0 --to v2.0.0
```

Available options:
- `--repo`: The GiHub repository to compare the releases from.
- `--token`: The GitHub token to use for the requests. _(Optional, defaults to none)_
- `--from`: The base release to compare from.
- `--to`: The release to compare to.
- `--output`: The output directory to download releases into. _(Optional, defaults to `./releases/`)_
- `--help`: Display the help message.
- `--version`: Display the version of the script.

## Installation

Just download the binary from the [releases page](https://github.com/WarningImHack3r/npm-stats-comparator/releases) and you're good to go!
