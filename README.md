# Decentralized Timeline

Developed for the 1st Semester's [Large Scale Distributed Systems class (M.EIC004)](https://sigarra.up.pt/feup/en/UCURR_GERAL.FICHA_UC_VIEW?pv_ocorrencia_id=501934) of the Master in Informatics and Computing Engineering (December 2022) 

This project explores the creation of a decentralized timeline service that harvests peer-to-peer and edge devices. 

- Users have an identity and publish small text messages in their local machine, forming a local timeline. 
- They can subscribe to other user’s timelines and will help to store and forward their content. 
- Remote content is available when a source or source subscriber is online and can forward the information. 
- Information from subscribed sources can be ephemeral and only stored and forwarded for a given time period.

## Implementation
The inital objective was to develop both a command line application and a frontend to interface with it in a more user-friendly way.

Only the command line interface is fully functional.

### Technologies
- Go
- LibP2P
- React.js

### Architectural aspects
- The network is ready to accept `peer` nodes when at least a `bootstrap` node is active; `bootstrap` nodes can be initialized using the application command line client
- `Peer` nodes can then be launched and interact with other `peer` nodes, posting small messages and subscribing to other users timelines
- Each `peer` is uniquely identified on the network, associating an username to each `peer` node, through the use of a Distributed Hash Table (`DHT`)
- Each username has a list of posted message identifiers saved on the `DHT`.
- Each message, uniquely identified, has a list of `peer` nodes that can have it in persistent storage, from which another node can retrieve it (named `provider` in this context).
- When a `peer` subcribes messages from a new username it fetches the posted messages list from the `DHT` and then queries it for `providers` for each message. It proceeds to retrieve them from available providers, and announces itself as a provider for the successfully retrieved messages.
- A faster delivery system was also put in place using a `publish-subriber` messge mechanism. Each username generates a `topic` where it also posts the new messages. Subscribed `peers` parse receive the new message in the `topic` and announce themselves as `providers` for that message.

The `announce` and `publish-substriber` mechanisms are provided by the `DHT` implementation used ([LibP2P](https://libp2p.io/), formerly part of [IPFS](https://ipfs.tech/)).

## Contributors
- Carlos Gomes ([@carlosmgomes](https://github.com/carlosmgomes))
- José Costa ([@Sirze01](https://www.github.com/Sirze01))
- Pedro Silva ([@PedroJSilva2001](https://github.com/PedroJSilva2001))
- Sérgio Estêvão ([@SergioEstevao11](https://github.com/SergioEstevao11))

