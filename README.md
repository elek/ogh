# Ozone Github Helper

Ozone GH is a simple script to make it easier the development of [Apache Hadoop Ozone](https://hadoop.apache.org/ozone)

## Usage

### Print out ready pull requests

```
> ogh review

+-----+--------------+----------------------------------------------------+------------------------+----------------+
| ID  |    AUTHOR    |                      SUMMARY                       |      PARTICIPANTS      |     CHECK      |
+-----+--------------+----------------------------------------------------+------------------------+----------------+
| 648 | >avijayanhwx | HDDS-3117. Recon throws InterruptedException while | ✓swagl                 | _______ ______ |
| 578 | >elek        | HDDS-3053. Decrease the number of the chunk writer | ✓adoro                 | _______ ______ |
| 622 | >adoroszlai  | HDDS-3113. Add new Freon test for putBlock         | ✓elek                  | _______ ______ |
| 551 | >adoroszlai  | HDDS-2717. Handle chunk increments in datanode     | bshas,adoro,arp7,lokes | _______ ______ |
| 618 | >captainzmc  | HDDS-2911. Fix lastUsed and stateEnterTime value i |                        | _______ ______ |
| 555 | >elek        | HDDS-3023. Create Freon test to test isolated Rati |                        | _______ ______ |
| 582 | >smengcl     | HDDS-3047. ObjectStore#listVolumesByUser and Creat | xiaoy                  | _______ ______ |
| 399 | >cxorm       | HDDS-2424. Add the recover-trash command server si | cxorm,bhara,maoba      | _______ ...... |
+-----+--------------+----------------------------------------------------+------------------------+----------------+
```

Legend: 
 * `[C]` means a conflict
 *  Participants can be prefixed with a review flag (✓ approved, ✕ change requested)
 * The last column show the results of the checks
   * `_` means a passed build
   * `.` means a missing build
   * `%` means an in-progress builds
   * any letter (eg. `b`,`c`) means a failing test (`b` -> build, `u` -> unit test ,etc). 
   * The second part of the checks display all the integrations tests.

### Print out all the available pull requests

```
ogh pr
+-----+---------------+----------------------------------------------------+---------------------------------+----------------+
| ID  |    AUTHOR     |                      SUMMARY                       |          PARTICIPANTS           |     CHECK      |
+-----+---------------+----------------------------------------------------+---------------------------------+----------------+
| 430 | >cxorm        | HDDS-2817. Fix listing buckets for setting --prefi | ✕githu,cxorm,smeng              | _______ ______ |
| 649 | >bharatviswa5 | [WIP]HDDS-3120. Freon work with OM HA.             |                                 | ______a _____o |
| 623 | >supratimdeka | HDDS-2941. file create : create key table entries  | mukul                           | _____u_ f_____ |
| 524 | >iamabug      | HDDS-2797. beyond/RunningWithHDFS.md translation   | ✕cxorm                          | _______ ______ |
| 520 | >iamabug      | HDDS-2793. concept/Datanodes.md translation        | iamab,xiaoy                     | ______a ______ |
| 525 | >iamabug      | HDDS-2798. beyond/Containers.md translation        | ✕cxorm,iamab                    | ______a f__h__ |
| 648 | >avijayanhwx  | HDDS-3117. Recon throws InterruptedException while | ✓swagl                          | _______ ______ |
| 608 | >sodonnel     | HDDS-3084 - Extended Network Topology Robot tests  | ✕adoro,✕githu,elek,sodon       | _______ ______ |
| 645 | >runzhiwang   | HDDS-3130. Add jaeger trace span in s3gateway      |                                 | _____u_ ______ |
| 578 | >elek         | HDDS-3053. Decrease the number of the chunk writer | ✓adoro                          | _______ ______ |
| 622 | >adoroszlai   | HDDS-3113. Add new Freon test for putBlock         | ✓elek                           | _______ ______ |
| 551 | >adoroszlai   | HDDS-2717. Handle chunk increments in datanode     | lokes,bshas,adoro,arp7          | _______ ______ |
| 643 | >hanishakoner | HDDS-2339. Add OzoneManager to MiniOzoneChaosClust |                                 | _____u_ ______ |
```

### Print out latest builds on master

```
ogh builds master

+-----+----------------------+--------------+---------------------+--------+----------------------------------------------------+----------------+
| ID  |       CREATED        |   WORKFLOW   |        REPO         | BRANCH |                       COMMIT                       |     CHECKS     |
+-----+----------------------+--------------+---------------------+--------+----------------------------------------------------+----------------+
| 579 | 2020-03-08T04:04:13Z | build-branch | apache/hadoop-ozone | master | HDDS-3089. TestSCMNodeManager intermittent crash ( | _______ ______ |
| 578 | 2020-03-07T07:35:10Z | build-branch | apache/hadoop-ozone | master | HDDS-3075. Fix ScmCli exception message when conta | _______ ______ |
| 576 | 2020-03-06T22:17:09Z | build-branch | apache/hadoop-ozone | master | HDDS-3071. Datanodes unable to connect to recon in | _______ ______ |
| 574 | 2020-03-06T17:37:19Z | build-branch | apache/hadoop-ozone | master | HDDS-3132. NPE when create RPC client. (#646)      | _____u_ f__h__ |
| 571 | 2020-03-06T15:53:19Z | build-branch | apache/hadoop-ozone | master | HDDS-3072. SCM scrub pipeline should be started af | _______ ______ |
| 570 | 2020-03-06T14:45:47Z | build-branch | apache/hadoop-ozone | master | HDDS-3131. Disable TestMiniChaosOzoneCluster (#644 | _______ ______ |
```

### Configuration

If you have ever used `hub` or `gh`, you don't need to set the GITHUB_TOKEN:

Github token can be set by

 * setting `GITHUB_TOKEN` environment variable
 * Setting it in `.config/hub` used by `hub`
 * Setting it in `.config/gh/config.yaml`

Use `OGH_CACHE` environment variable to defined a file (prefix) for cache files. Results will be cached for a short time (<5 mins)

```
OGH_CACHE=/tmp/ogh ogh pr
 ```

### Interactive

You can use it as an interactive command with `fzf`

```
ogh review | fzf --reverse | awk '{print $2}' | xargs -n1 ogh
```

## Install

OSX:

```
brew install elek/brew/ogh
```

Linux

```
go get github.com/elek/ogh
```
