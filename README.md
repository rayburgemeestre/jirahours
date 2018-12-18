## Install

- Compile yourself or copy `jirahours` into your path. (You can download it here: https://cppse.nl/jirahours)
- Copy `example/jirahours.yaml` to `~/jirahours.yaml` and modify it.

## Example usage

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
- `jirahours generate` produces an `submit_hours.sh` script you can invoke to submit the Tempo hours to Jira.

Each step you can manually inspect and change stuff. For example `issues.txt` might need some polishing, or `submit_hours.sh` needs to be double checked if everything looks good.

Executing `submit_hours.sh` will make calls to jirahours, to the following command:

- `jirahours submit` which can be used to send worklogs (one by one) to Jira. See `--help` for more info.

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

The semi-colon "comments" out the date, but it will be used to fetching the issues by min and max date.

## Other commands

- `jirahours fetch` to git fetch --all on each repository.

