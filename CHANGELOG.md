# Release Notes

## v1.1.2 / 2022-03-10
- downgrade ugorji, see: https://github.com/ugorji/go/issues/369

## v1.1.1 / 2022-03-09
- update ugorji/go/codec version

## v1.1.0 / 2022-03-08
- sync with gitlab
- update go version

## v1.0.1 / 2020-11-24
- migrated to GitHub

## v1.0.0 / 2020-01-07
- release version v1.0.0
- add verify script and ci job 

## 0.2.0 / 2019-11-08
- update to use gomods
- update ugorji library for go modules compatibility

## 0.1.4 / 2019-10-15
- added XOR(UUID) method that calculates the bitwise XOR result of two UUIDs,
making a new, valid UUID from them

## 0.1.3 / 2019-09-26
- added NewTime(time.Time) and TimeUUIDToTime() methods to support TimeUUID
- added Timestamp() and Time() helper methods

## 0.1.2 / 2019-03-22
- added FromHashLike(string) and UUID.HashLike() methods
- added Next() method

## 0.1.1 / 2018-08-21
- add sql interface support
- add gitlab ci to run test automatically

## 0.1.0 / 2018-08-16
- project created
