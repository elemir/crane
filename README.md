# crane ![CircleCI](https://img.shields.io/circleci/project/github/elemir/crane.svg) ![license](https://img.shields.io/github/license/elemir/crane.svg)
## description
crane -- small utility for debugging containerized application. It creates special debug container from prepared image which:

* uses IPC, PID и network namespaces from debugged container (works only for running container, may be skipped)
* mount its filesystem into special path /cont

