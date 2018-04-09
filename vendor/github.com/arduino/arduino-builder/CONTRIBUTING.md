## Submitting an issue

It's crucial that you provide full verbose debug output, attached or via links to sites like http://gist.github.com/ or http://pastebin.com/

If you're using the Arduino IDE:
* turn on verbose compilation from "File > Preferences" dialog
* copy the full command line
* add parameter `-debug-level=10` before the sketch path
* manually re-run the builder in a terminal or command prompt

If you are using this tool from the command line, ensure you have specified both `-verbose` and `-debug-level=10`

## Submitting a pull request

We are glad you want to contribute with code: that's the best way to help this software.

Your contribution is adding or modifying existing behaviour. We are used to [Test Driven Development](https://en.wikipedia.org/wiki/Test-driven_development): please add one or more tests that prove that your contribution is good and is working as expected.

Be sure to run the provided `fmt_fix_vet` script before each commit: it ensures your code is properly formatted.

Also, for your contribution to be accepted, everyone of your commits must be "Signed-off". This is done by commiting using this command: `git commit --signoff`

By signing off your commits, you agree to the following agreement, also known as [Developer Certificate of Origin](http://developercertificate.org/): it assures everyone that the code you're submitting is yours or that you have rights to submit it.

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```
