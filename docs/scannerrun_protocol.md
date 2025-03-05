# Scanner Run protocol

## Background

Heureka needs to keep track of scanner runs for two primary reasons:

1. For compliance reasons Heureka needs to track that scanner runs were
executed.
2. Heureka needs information for detected security vulnerabilities to detect the
presence, and absence of security vulnerabilities.

Two lifetime events are relevant for tracking scanner runs:

1. The start of a scanner run.
2. The end of a scanner run.

Further, every finding of a scanner run has to be associated with the correct
scanner run, to enable us to detect the absence/mitigation of a vulnerability
automatically.

There is an asymmetry concerning the presence and absence of security
vulnerabilities: Every reported security vulnerability is a positive finding
and can be reported immediately on detection. Even if the scanner run fails to
end for any reason, this is true. The absence of a vulnerability on the other
side can only be detected if a scanner run ends successfully and has not
detected a vulnerability which was detected in a prior scanner run.  (Actually
detecting the absence of a security vulnerability is more complicated and the
absence of a prior seen vulnerability is not a guarantee the vulnerability is
gone, but this is out of the scope for the current document.)

## The protocol

The participants for the protocol are one instance of Heureka (H) and 1..N
instances of one type of Scanner (S).

To associate different instances of scanner runs with each other each scanner
has a tag which can be arbitrarily chosen as long as it does not clash with
another scanner working with the same Heureka instance. The tag stays constant between all run

The scanner is the driver of the protocol.

### Create a new scanner run

S creates a new random UUIDv4 to identify the current scanner run. The random
UUIDv4 is used to identify the scanner run and thus must be kept until the
completion or failure of the scanner run.

Afterwards S calls the following GraphQL query:

    input ScannerRunInput {
       uuid:   String
        tag:    String
    }

    createScannerRun(input: ScannerRunInput!): Boolean!


### Complete a scanner run

If the scanner run completed successfully, S calls the following GraphQL query:

    completeScannerRun(uuid: String!): Boolean!

### Fail a scanner run

If the scanner run fails for any reason, S calls the following GraphQL query:

    failScannerRun(uuid: String!, message: String!): Boolean!
