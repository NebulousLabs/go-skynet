# Changelog

## [2.0.1]

### Changed

- Fixed single-file directory uploads being uploaded as files instead of
  directories.

## [2.0.0]

### Changed

- This SDK has been updated to match Browser JS and require a client. You will
  first need to create a client and then make all API calls from this client.
- Connection options can now be passed to the client, in addition to individual
  API calls, to be applied to all API calls.
- The `defaultPortalUrl` string has been renamed to `defaultSkynetPortalUrl` and
  `defaultPortalUrl` is now a function.
- Fix a bug where the Content-Type was not set correctly.

## [1.1.0]

*Prior version numbers skipped to maintain parity with API.*

### Added

- Generic `Upload` and `Download` functions
- Common Options object
- API authentication
- Encryption

### Changed

- Some upload bugs were fixed.

## [0.1.0]

### Added

- Upload and download functionality.
