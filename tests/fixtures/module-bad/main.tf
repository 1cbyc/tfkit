terraform {
  backend "local" {}
}

resource "terraform_data" "bad" {
  provisioner "local-exec" {
    command = "false"
  }
}
