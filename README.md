Helper program to fetch [Euraxess](https://euraxess.ec.europa.eu/) job offers from a particular URI and post them to MediaWiki.

To quickly deploy to the server, update `Makefile` for your server location, binary and templates destinations, then run (assuming SSH is up and configured):

```
$ make deploy
```

Run it on the server like this:

```
$ euraxess-pull -mwuri https://ims.ut.ee/ -name "UserName" -pass "pass" -page Job_Offers -tmpl offers.tmpl -uri "https://euraxess.ec.europa.eu/jobs/search/country/estonia-1069"
```
