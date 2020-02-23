
use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize, Debug)]
struct Cartesian {
    x: f64,
    y: f64,
}

#[derive(Serialize, Deserialize, Debug)]
struct Polar {
    r: f64,
    theta: f64,
}

pub fn run(){
    let p = Cartesian {x:3.0, y:2.0};
    let client = reqwest::blocking::Client::new();
    let res = client.post("http://127.0.0.1:8001/cartesian-to-polar")
        .json(&p)
        .send()
        .unwrap();

    println!("{:?}", res);
    let polar: Polar = res.json().unwrap();
    println!("{:?}", polar);
}
