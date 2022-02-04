# 3rd-party-cookie

3rd party Cookie を用いてトラッキングするサンプルです

## 準備

- `tracker` が 127.0.0.1 に向くように hosts をいじる
- mkcert 等を利用して `localhost` の証明書ペア `client/localhost-key.pem` （秘密鍵）、 `client/localhost.pem` （証明書） と `tracker` の証明書ペア `tracker/tracker-key.pem` （秘密鍵）、 `tracker/tracker.pem` （証明書）を用意する

## 検証

1. `make serve-client` `make serve-tracker` を叩き、クライアント localhost:8080 とトラッカー tracker:9090 を起動する
2. https://localhost:8080 へアクセスしてみる。 tracker の 3rd party Cookie `identifier` が作成されることが確認できる
3. https://tracker:9090/me へアクセスしてみる。収集されたデータを確認できる
