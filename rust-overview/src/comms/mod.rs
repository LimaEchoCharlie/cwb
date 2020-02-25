mod coordinate;

pub fn run_http(){
    let p = coordinate::Cartesian {x:3.0, y:2.0};
    let client = reqwest::blocking::Client::new();
    let res = client.post("http://127.0.0.1:8001/cartesian-to-polar")
        .json(&p)
        .send()
        .unwrap();

    println!("HTTP response: {:?}", res);
    let polar: coordinate::Polar = res.json().unwrap();
    println!("HTTP polar: {:?}", polar);
}

pub fn run_coap(){
    // serialize cartesian coordinate
    let p = coordinate::Cartesian {x:3.0, y:2.0};
    let serialized = serde_cbor::to_vec(&p).unwrap();

    // create request
    let mut request = coap::CoAPRequest::new();
    request.set_method(coap::Method::Post);
    request.set_path("/cartesian-to-polar");
    request.message.set_content_format(coap::message::packet::ContentFormat::ApplicationCBOR);
    request.message.set_payload(serialized);

    // send request
    let client = coap::CoAPClient::new("127.0.0.1:5688").unwrap();
    client.send(&request).unwrap();

    // receive response and deserialize polar coordinate
    let response = client.receive().unwrap();
    println!("COAP response: {:?}", response);
    let polar: coordinate::Polar = serde_cbor::from_slice(&response.message.payload).unwrap();
    println!("COAP polar: {:?}", polar);
}
