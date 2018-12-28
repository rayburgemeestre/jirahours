## Install

- Compile yourself or copy `jirahours` into your path. (You can download it here: https://cppse.nl/jirahours)
- Copy `example/jirahours.yaml` to `~/jirahours.yaml` and modify it.

## Build

- Make sure you have go installed. I.e. `apt install golang` on Ubuntu.
- Checkout the source: `go get github.com/rayburgemeestre/jirahours` (Might take a while, pulls in also dependencies)
- Move to the checkout directory: `cd $HOME/go/src/github.com/rayburgemeestre/jirahours` 
- Run `go build` and you should have a `jirahours` in your current directory
- Run `cp -prv jirahours ~/.bin/` or something similar, to get it into your path.

## Usage

	trigen@zenbook:~/projects/jirahours[master]> jirahours 
    jirahours - because logging hours in jira is boring

    Usage:
      jirahours [command]

    Available Commands:
      fetch       Fetch all remotes on all repositories
      generate    Generate a bash script to submit jira hours
      help        Help about any command
      issues      Read in a dates file and gather all relevant git commit messages for the min and max date found in this file.
      submit      Submit hours to jira based on credentials specified in your config.
      syncback    Sync back already existing worklogs for given date range from Jira Tempo hours
      version     Print the version number of jirahours

    Flags:
          --config string   config file (default is $HOME/jirahours.yaml)
      -h, --help            help for jirahours

    Use "jirahours [command] --help" for more information about a command.

## Example

The typical workflow is the following.
Edit in your current directory a file named `dates.txt`, for example:

    2018-12-03
    2018-12-04
    2018-12-05
    2018-12-06
    2018-12-07
    2018-12-10
    2018-12-11
    2018-12-12
    2018-12-13
    2018-12-14
    2018-12-17
    2018-12-18

These dates will be the days that you have worked and you wish to generate jira hours for (later)

- `jirahours issues` produces an `issues.txt` with all commits from repositories you specified in your config that go from the min- and max dates in this `dates.txt`.
- `jirahours syncback` produces an `existing_tempo_hours.txt` with all existing worklog entries already in Tempo, typically meetings etc. 
- `jirahours generate` produces an `submit_hours.sh` script you can invoke to submit the Tempo hours to Jira. This will read `issues.txt`, `dates.txt` and `existing_tempo_hours.txt` to construct a script.

Each step you can manually inspect and change stuff. For example `issues.txt` might need some polishing before doing the `generate` step,
 or `submit_hours.sh` needs to be double checked until you are confident enough to run the bash script.

Executing `submit_hours.sh` will make calls to `jirahours` as well, to the following command to be specific:

- `jirahours submit` which can be used to send worklogs (one by one) to Jira. See `--help` for more info.

Note all above commands have parameters, and support `--help` or `jirahours help <command>`.


## Using a different range for dates.txt

Say your first day in `dates.txt` is a Monday, but you also worked on the weekend before, you might want to add the following hints in your `dates.txt`:

    ; 2018-12-01
    2018-12-03
    2018-12-04
    2018-12-05
    2018-12-06
    2018-12-07
    2018-12-10
    2018-12-11
    2018-12-12
    2018-12-13
    2018-12-14
    2018-12-17
    2018-12-18
    ; 2018-12-25

The semi-colon "comments" out the date, but it will be used when fetching the issues by min and max date.

## Other commands

- `jirahours fetch` to invoke `git fetch --all` on each repository. This can be useful in case you made changes on a different machine, to make your local copies aware of the commits.

## Why?

Well this project was just an excuse to get better at golang, since I already had this above logic in some bash scripts writing it was pretty easy.
The advantage is that the logic seems to be stable, I've been using this for quite a while.

## DISCLAIMER

The code is still a bit messy, hopefully I will get around to refactoring it a little bit.

## LICENSE

    This Source Code Form is subject to the terms of the Mozilla Public                                                                  
    License, v. 2.0. If a copy of the MPL was not distributed with this                                                                  
    file, You can obtain one at http://mozilla.org/MPL/2.0/.                                                                             
  
See LICENSE file for the complete text.

