# crane ![CircleCI](https://img.shields.io/circleci/project/github/elemir/crane.svg) ![license](https://img.shields.io/github/license/elemir/crane.svg)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Felemir%2Fcrane.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Felemir%2Fcrane?ref=badge_shield)
## description
crane -- small utility for debugging containerized application. It creates special debug container from prepared image which:

* uses IPC, PID и network namespaces from debugged container (works only for running container, may be skipped)
* mount its filesystem into special path /cont



## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Felemir%2Fcrane.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Felemir%2Fcrane?ref=badge_large)