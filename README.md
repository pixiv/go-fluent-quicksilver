# go-fluent-quicksilver

This project is under developing now.

## Features

- [x] forward output
  - [x] buffering queues
  - [x] sending in [PackedForwardMode](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1#packedforward-mode)
  - [ ] heartbeat
  - [ ] receiving ACK response
  - [x] data marshaling using [tinylib/msgp](https://github.com/tinylib/msgp)
  - [ ] file buffer
- [ ] forward input
- [ ] file input
- [ ] logger

## License

```
Copyright 2018 pixiv Technologies Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
