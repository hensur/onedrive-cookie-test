# Onedrive Cookie Auth Proof of Concept

Some Microsoft Accounts (mostly university accounts) are in an unmanaged state. This happens if no DNS validation has been done. Microsoft allows login and full use of all webservices in this state, however, the API can't be used with these Accounts. Tools like rclone rely on the API to upload files to onedrive and cant authenticate. More details in this issue: [rclone #1975](https://github.com/ncw/rclone/issues/1975)

Apparently there is another way to access onedrive on an unmanaged Account. It is possible to use the WebDAV Endpoint which requires a valid cookie for authentication. Tools like davfs2 are able to mount a share by cookie authentication. More details can be foud in this [blog post](https://shui.azurewebsites.net/2018/01/13/mount-onedrive-for-business-on-headless-linux-vps-through-webdav/)

This repo contains a go rewrite of this [python implementation](https://github.com/yulahuyed/test) to aquire the cookie.
I'd like to use my university with rclone. After getting a working cookie, I'd like to extend rclone with this feature. Which hopefully won't be that hard, since they already have a webdav connector.