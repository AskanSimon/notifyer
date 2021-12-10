# Notifyer

With this simple notifyer, you can periodically login to one of your web accounts and if there are important updates, you will get an email (or notification on your smartphone).

Here a Cronjob and Chromedp is used, to login to your account. With regular expression the content after the login is filtered. If there are updates, an email is send to a new gmail adress. The corresponding gmail app on your smartphone for is configured to ring immediately on every new email. Its okey, when you use the gmail app just for this adress.

### Requirements 

You will need a login (without captcha), a new gmail adress, a host to send emails and you will need to configure chromedp in go, so it can read your specific website. You will find an example in main.go, what this can look like.


