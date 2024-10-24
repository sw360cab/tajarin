# Tajarin

An handy application collecting validators' credentials (hostname, pub key, address, p2p node key) to generate and distribute a shared Genesis file and partial Gno configuration items (_persistent_peers_).

The goal of the application is the possibility to create and distibute a genesis file (plus other configuration) to be shared among the validators that will partecipate in the network of Gno nodes. It avoids the need to agree, build and distribute manually the genesis file.

Based on pub/sub model and event driven paradigm.

## How it works

The application has two tiers:

- one publisher
- one or more subscribers (validators)

Validators looking for a Genesis file subscribe to the service by providing a set of data.
Upon receving a predefined number of subscrptions the service will use the given information to generate

- the genesis file
- any configuration key

The generated data will be returned to the subscriber in a JSON format:

```json
{
  'genesis': <genesis_file_content>,
  'config': {
    persistent_peers: <value>
    <config_key_1>: ...,
    <config_key_2>: ...
  }
}
```

### Subscriber

Connect to the service providing the following data:

- name (`-name`)
- pub key (`-pub-key`)
- address (`-address`)
- p2p node key (`-p2p-key`)
- p2p node public address or dns name (`-p2p-host`)
- p2p node port (`-p2p-port`) [optional, default: 26656]

The previous value are mandatory (if not specified differently) to subscribe and any miss will result in a rejected request.
The subscriber will be notified with request data when the publisher has an answer ready.

**NOTE**: At the moment the connection is intended to be on the same network on port `1900`

### Publisher

The collector is spawned with an argument, an integer number representing the number of validators that should be syncronized
within the netwrok of nodes.
It will act in the following wait:

- wait until up to # validators subscribe to the service
- upon reaching the number of valid validators subscribed
  - combine the information into the genesis file
  - generate conifguration snippets
  - broadcast the given data to every subscriber in JSON format

## Origin of the name

:spaghetti:

[Tajarin](https://en.wikipedia.org/wiki/Tagliolini) is a special type of Italian pasta coming from Turin and the surrounding Piedmont region, made of egg dough. It has the same shape of Spaghetti but is more it is more similar to Fettuccine.

## Future work

- generate secrets within Tajarin itself, optionally and via argument param
