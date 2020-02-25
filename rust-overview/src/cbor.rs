use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize, Debug)]
struct Point {
    x: i32,
    y: i32,
}

pub fn run() {
    let point = Point { x: 1, y: 2 };

    let serialized = serde_cbor::to_vec(&point).unwrap();
    println!("cbor serialized = {:?}", serialized);

    let deserialized: Point = serde_cbor::from_slice(&serialized).unwrap();
    println!("cbor deserialized = {:?}", deserialized);
}
