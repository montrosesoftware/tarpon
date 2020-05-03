# Tarpon

Golang simple, minimalist, secure and fast signalling server using WebSockets.

Its common use case is signalling in WebRTC video conferencing, e.g. establishing sessions
between peers, but can also be used to easily add group chat features.

## Basics

**Room** is a group of **peers** who can communicate with each other by either sending
direct **messages** or broadcasting.

Direct messages can't be read by other **peers**. **Peers** need to be registered prior
joining the given room.

The whole state is stored in memory, so it gets lost on server restart.

**Messages** contain _id_ to match associated request/response messages and _senderId_
to securely identify the sender of the given message.

Any sender who sends too many messages will be disconnected by **Tarpon**. This should prevent
simple DOS attacks from malicious senders.
