# npm-stats-comparator

Compare a few statistics between two GitHub releases of an NPM package. **For now, only the LOCs comparison is available.**  
Includes the intermediate releases between the two specified releases.

_Inspired by [this tweet](https://twitter.com/denlukia/status/1772818790415225202) by [@denlukia](https://github.com/denlukia)._

[![Demo](https://github.com/WarningImHack3r/npm-stats-comparator/assets/43064022/5e822a1d-3d1c-4e6e-b381-4751220cac59)](https://twitter.com/probably_coding/status/1774934048114164069)
:---:
*Demo (click the image to watch the video!)*

## Usage

```bash
$ ./npm-stats-comparator --repo user/repo --from v1.0.0 --to v2.0.0
```

Available options:
- `--repo`: The GitHub repository to compare the releases from.
- `--token`: The GitHub token to use for the requests. _(Optional, defaults to none)_
- `--from`: The base release to compare from.
- `--to`: The release to compare to.
- `--output`: The output directory to download releases into. _(Optional, defaults to `./releases/`)_
- `--help`: Display the help message.
- `--version`: Display the version of the script.

## Installation

Just download the binary from the [releases page](https://github.com/WarningImHack3r/npm-stats-comparator/releases) and you're good to go!
