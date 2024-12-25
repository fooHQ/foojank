# Installation

## Setup NATS server

### Install NATS server

```
$ curl -sf https://binaries.nats.dev/nats-io/nats-server/v2@latest | sh
$ curl # TODO copy nats-server.conf from the repository
```

## Prepare nsc environment

nsc is an official tool that can be used to manage operators, account and users on a NATS server. The following tutorial will
lead you through the process of setting up the initial environment.

Never initialize nsc environment on the server.

### Create a new operator

```
$ nsc init --name smooth-operator
[ OK ] created operator smooth-operator
[ OK ] created system_account: name:SYS id:ACUKRQEA7OIYWJSHPJWZGJHZJEL3LEDFMRLGSL5USCNJ7B66GGVLIKXO
[ OK ] created system account user: name:sys id:UBK7LTT34HJHBWID3SBHXVYCLCSU57NXQPBKKOZ4PFV4XH7M7O3DQ345
[ OK ] system account user creds file stored in `~/.local/share/nats/nsc/keys/creds/smooth-operator/SYS/sys.creds`
[ OK ] created account smooth-operator
[ OK ] created user "smooth-operator"
[ OK ] project jwt files created in `~/.local/share/nats/nsc/stores`
[ OK ] user creds file stored in `~/.local/share/nats/nsc/keys/creds/smooth-operator/smooth-operator/smooth-operator.creds`
> to run a local server using this configuration, enter:
>   nsc generate config --mem-resolver --config-file <path/server.conf>
> then start a nats-server using the generated config:
>   nats-server -c <path/server.conf>
all jobs succeeded
``` 

### Generate NATS resolver configuration

```
$ nsc generate config --nats-resolver > nats-resolver.conf
$ scp nats-resolver.conf user@example.com:~/
```

### Create account and users

#### Create account

```
$ nsc add account -n ACCOUNT-001
[ OK ] generated and stored account key "ABWE6VNFVG3S42QDCGLRQN6ZMBNN4Q3LRE2VAAIXQKOH6CZTP5S44TWU"
[ OK ] added account "ACCOUNT-001"
```

#### Create manager user

```
$ nsc add user --account ACCOUNT-001 --name johndoe
[ OK ] generated and stored user key "UD4TTBCJXVLYEL5MNV2QJDBI5CLPPKPBL2L37YI7ZAND5NGSX6KTZ373"
[ OK ] generated user creds file `~/.local/share/nats/nsc/keys/creds/smooth-operator/ACCOUNT-001/johndoe.creds`
[ OK ] added user "johndoe" to account "ACCOUNT-001"
```

#### Create agent user

```
$ U=agent001; nsc add user --account ACCOUNT-001 --name $U \
    --allow-sub '_INBOX_'$U'.>' \
    --allow-sub '$SRV.PING' \
    --allow-sub '$SRV.PING.'$U \
    --allow-sub '$SRV.PING.'$U'.*' \
    --allow-sub '$SRV.INFO' \
    --allow-sub '$SRV.INFO.'$U \
    --allow-sub '$SRV.INFO.'$U'.*' \
    --allow-sub '$SRV.STATS' \
    --allow-sub '$SRV.STATS.'$U \
    --allow-sub '$SRV.STATS.'$U'.*' \
    --allow-sub $U'.RPC' \
    --allow-sub $U'.*.DATA' \
    --allow-sub $U'.*.STDIN' \
    --allow-pub '_INBOX_'$U'.>' \
    --allow-pub '$JS.API.STREAM.INFO.OBJ_'$U \
    --allow-pub '$JS.API.DIRECT.GET.OBJ_'$U'.$O.'$U'.M.*' \
    --allow-pub '$JS.API.CONSUMER.CREATE.OBJ_'$U'.*.$O.'$U'.C.*' \
    --allow-pub '$JS.API.CONSUMER.DELETE.OBJ_'$U'.*' \
    --allow-pub $U'.*.STDOUT' \
    --allow-pub-response 1
[ OK ] set max responses to 1
[ OK ] added pub "_INBOX_agent001.>"
[ OK ] added pub "$JS.API.STREAM.INFO.OBJ_agent001"
[ OK ] added pub "$JS.API.DIRECT.GET.OBJ_agent001.$O.agent001.M.*"
[ OK ] added pub "$JS.API.CONSUMER.CREATE.OBJ_agent001.*.$O.agent001.C.*"
[ OK ] added pub "$JS.API.CONSUMER.DELETE.OBJ_agent001.*"
[ OK ] added pub "agent001.*.STDOUT"
[ OK ] added sub "_INBOX_agent001.>"
[ OK ] added sub "$SRV.PING"
[ OK ] added sub "$SRV.PING.agent001"
[ OK ] added sub "$SRV.PING.agent001.*"
[ OK ] added sub "$SRV.INFO"
[ OK ] added sub "$SRV.INFO.agent001"
[ OK ] added sub "$SRV.INFO.agent001.*"
[ OK ] added sub "$SRV.STATS"
[ OK ] added sub "$SRV.STATS.agent001"
[ OK ] added sub "$SRV.STATS.agent001.*"
[ OK ] added sub "agent001.RPC"
[ OK ] added sub "agent001.*.DATA"
[ OK ] added sub "agent001.*.STDIN"
[ OK ] generated and stored user key "UAIYMP57FZIFK546ZH6UL7XSBP6ICGNC4NI64LG66AGB764H6WQODE6C"
[ OK ] generated user creds file `~/.local/share/nats/nsc/keys/creds/smooth-operator/ACCOUNT-001/agent001.creds`
[ OK ] added user "agent001" to account "ACCOUNT-001"
```

### Generate credential file

```
$ nsc generate creds --account ACCOUNT01 --name agent001
```
