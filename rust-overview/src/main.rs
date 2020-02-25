mod json;
mod cbor;
mod http;

fn main() {
    json::run();
    http::run();
    cbor::run();

}
