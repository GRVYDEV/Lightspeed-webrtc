
<p align="center">
<a  href="https://github.com/GRVYDEV/Lightspeed-webrtc">
    <img src="images/lightspeedlogo.svg" alt="Logo" width="150" height="150">
</a>
</p>
  <h1 align="center">Project Lightspeed WebRTC</h1>
<div align="center">
  <a href="https://github.com/GRVYDEV/Lightspeed-webrtc/stargazers"><img src="https://img.shields.io/github/stars/GRVYDEV/Lightspeed-webrtc" alt="Stars Badge"/></a>
<a href="https://github.com/GRVYDEV/Lightspeed-webrtc/network/members"><img src="https://img.shields.io/github/forks/GRVYDEV/Lightspeed-webrtc" alt="Forks Badge"/></a>
<a href="https://github.com/GRVYDEV/Lightspeed-webrtc/pulls"><img src="https://img.shields.io/github/issues-pr/GRVYDEV/Lightspeed-webrtc" alt="Pull Requests Badge"/></a>
<a href="https://github.com/GRVYDEV/Lightspeed-webrtc/issues"><img src="https://img.shields.io/github/issues/GRVYDEV/Lightspeed-webrtc" alt="Issues Badge"/></a>
<a href="https://github.com/GRVYDEV/Lightspeed-webrtc/graphs/contributors"><img alt="GitHub contributors" src="https://img.shields.io/github/contributors/GRVYDEV/Lightspeed-webrtc?color=2b9348"></a>
<a href="https://github.com/GRVYDEV/Lightspeed-webrtc/blob/master/LICENSE"><img src="https://img.shields.io/github/license/GRVYDEV/Lightspeed-webrtc?color=2b9348" alt="License Badge"/></a>
</div>
<br />
<p align="center">
  <p align="center">
    A RTP -> WebRTC server based on Pion written in Go. This server accepts RTP packets on port 65535 and broadcasts them via WebRTC
    <!-- <br /> -->
    <!-- <a href="https://github.com/GRVYDEV/Lightspeed-webrtc"><strong>Explore the docs »</strong></a> -->
    <br />
    <br />
    <a href="https://github.com/GRVYDEV/Lightspeed-webrtc">View Demo</a>
    ·
    <a href="https://github.com/GRVYDEV/Lightspeed-webrtc/issues">Report Bug</a>
    ·
    <a href="https://github.com/GRVYDEV/Lightspeed-webrtc/issues">Request Feature</a>
  </p>
</p>



<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary><h2 style="display: inline-block">Table of Contents</h2></summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgements">Acknowledgements</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

<!-- [![Product Name Screen Shot][product-screenshot]](https://example.com) -->

This is one of three components required for Project Lightspeed. Project Lightspeed is a fully self contained live streaming server. With this you will be able to deploy your own sub-second latency live streaming platform. This particular repository takes RTP packets sent to the server and broadcasts them over WebRTC. In order for this to work the Project Lightspeed Ingest server is required to perfrom the FTL handshake with OBS. In order to view the live stream the Project Lightspeed viewer is required.


### Built With

* Pion
* Golang


<!-- GETTING STARTED -->
## Getting Started

To get a local copy up and running follow these simple steps.

### Prerequisites

In order to run this Golang is required. Installation instructions can be found <a href="https://golang.org/doc/install#download">here</a>

### Installation

Get the repo
   ```sh
   export GO111MODULE=on
   go get github.com/GRVYDEV/lightspeed-webrtc
   ```


<!-- USAGE EXAMPLES -->
## Usage

To run type the following command. The host ip argument is the local IP address of your machine. This is the IP that the server uses to listen for RTP Packets on
```sh
   lightspeed-webrtc --host-ip=XX.XX.XX.XX
   ```


<!-- _For more examples, please refer to the [Documentation](https://example.com)_ -->



<!-- ROADMAP -->
## Roadmap

See the [open issues](https://github.com/GRVYDEV/Lightspeed-webrtc/issues) for a list of proposed features (and known issues).



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request



<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.



<!-- CONTACT -->
## Contact

Garrett Graves - [@grvydev](https://twitter.com/grvydev)

Project Link: [https://github.com/GRVYDEV/Lightspeed-webrtc](https://github.com/GRVYDEV/Lightspeed-webrtc)



<!-- ACKNOWLEDGEMENTS -->
## Acknowledgements

* []()
* []()
* []()





<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/GRVYDEV/repo.svg?style=for-the-badge
[contributors-url]: https://github.com/GRVYDEV/repo/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/GRVYDEV/repo.svg?style=for-the-badge
[forks-url]: https://github.com/GRVYDEV/repo/network/members
[stars-shield]: https://img.shields.io/github/stars/GRVYDEV/repo.svg?style=for-the-badge
[stars-url]: https://github.com/GRVYDEV/repo/stargazers
[issues-shield]: https://img.shields.io/github/issues/GRVYDEV/repo.svg?style=for-the-badge
[issues-url]: https://github.com/GRVYDEV/repo/issues
[license-shield]: https://img.shields.io/github/license/GRVYDEV/repo.svg?style=for-the-badge
[license-url]: https://github.com/GRVYDEV/repo/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/GRVYDEV

