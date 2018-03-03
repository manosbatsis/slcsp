

## Overview

A solution to the [homework provided by AdHoc](https://github.com/adhocteam/homework/tree/master/slcsp), written in Go.

Please note this is a Java guy's very first Go program, an attempt to something useful for yours truly personally.
The code was written for an interview test with no real use for the recipient anyway, so I saw it as a nice
opportunity to give the language a try.

Thus, you will probably not find a professional sample of Go code here, especially since I tend to see things from
an OO perspective and have little knowledge of the language properties, idioms and conventions. This actually took
hours to write while googling for the absolute basics. Even worse, anything is possible without a test case!

## Assumptions

The code makes the following assumptions:

- The objective is to find, for each zip code, the SLCSP of the rate area it corresponds to
- The county code is irrelevant to the problem, as rate areas is what actually connects zip codes plan rates
- The relationship between zip codes and rate areas is considered many-to-one in relational terms
- Zip codes mapped to multiple rate areas are ignored, i.e. considered ambiguous and left blank per the instructions
- Zip codes for which less than two silver plans are discovered are to be given No SLCSP value in the resulting file
- An in-memory index is the reasonable performance trade-off, i.e. rather small input files are expected
- The `state` and `rate_area` columns in zips.scv and plans.csv compose the business key of a rate area (the tuple)
- All records in zips.csv and plans.csv are complete, i.e. without missing any column values
- No other input data validation or correction is necessary

Tp run this program you need to have Go installed in your path. Simply navigate to the slcsp folder in your command
line interface and execute the following:

## Run

```
go run slcsp.go
```

The program will produce the folowing output:

```
INFO:    2017/11/20 03:46:29 slcsp.go:201: Parsed 51541 records from zips.csv to 38804 zip codes and 477 rate areas
WARNING: 2017/11/20 03:46:29 slcsp.go:205: Note: zips.csv contained 3723 ambiguous zip codes (see trace or Appendix C in COMMENTS)
INFO:    2017/11/20 03:46:29 slcsp.go:377: Parsed 22239 records from plans.csv to 22239 plans
WARNING: 2017/11/20 03:46:29 slcsp.go:380: Note: plans.csv contained 7 unmapped areas (see trace or Appendix A in COMMENTS)
INFO:    2017/11/20 03:46:29 slcsp.go:352: Wrote 51 records to slcsp-modified.csv
WARNING: 2017/11/20 03:46:29 slcsp.go:355: Note: 10 zip codes had insufficient plan info, i.e. less than two plans (see trace or Appendix B in COMMENTS)
```

Appendix A, B, C: see the [COMMENTS](COMMENTS) file

## Original Instructions

### Calculate second lowest cost silver plan (SLCSP)

#### Problem

You have been asked to determine the second lowest cost silver plan (SLCSP) for
a group of ZIP Codes.

#### Task

You have been given a CSV file, `slcsp.csv`, which contains the ZIP Codes in the
first column. Fill in the second column with the rate (see below) of the
corresponding SLCSP. Your answer is the modified CSV file, plus any source code
used.

Write your code in your best programming language.

The order of the rows in your answer file must stay the same as how they
appeared in the original `slcsp.csv`.

It may not be possible to determine a SLCSP for every ZIP Code given. Check for cases
where a definitive answer cannot be found and leave those cells blank in the output CSV (no
quotes or zeroes or other text).

#### Additional information

The SLCSP is the so-called "benchmark" health plan in a particular area. It is
used to compute the tax credit that qualifying individuals and families receive
on the marketplace. It is the second lowest rate for a silver plan in the rate area.

For example, if a rate area had silver plans with rates of
`[197.3, 197.3, 201.1, 305.4]`, the SLCSP for that rate area would be `201.1`, since
it is the second lowest rate in that rate area.

A plan has a "metal level", which can be either Bronze, Silver, Gold, Platinum,
or Catastrophic. The metal level is indicative of the level of coverage the plan
provides.

A plan has a "rate", which is the amount that a consumer pays as a monthly
premium, in dollars.

A plan has a "rate area", which is a geographic region in a state that
determines the plan's rate. A rate area is a tuple of a state and a number, for
example, NY 1, IL 14.

There are two additional CSV files in this directory besides `slcsp.csv`:

* `plans.csv` -- all the health plans in the U.S. on the marketplace
* `zips.csv` -- a mapping of ZIP Code to county/counties & rate area(s)

A ZIP Code can potentially be in more than one county. If the county can not be
determined definitively by the ZIP Code, it may still be possible to determine
the rate area for that ZIP Code.

A ZIP Code can also be in more than one rate area. In that case, the answer is ambiguous
and should be left blank.

We will want to compile your code from source and run it, so please include the
complete instructions for doing so in a COMMENTS file.
