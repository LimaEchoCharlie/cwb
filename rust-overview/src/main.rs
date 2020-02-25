//mod json;
//mod cbor;
mod comms;

fn main() {
//    json::run();
//    cbor::run();
    comms::run_http();
    comms::run_coap();
}
