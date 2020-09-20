resource "aws_dynamodb_table" "tf_provider_data_processing_feed" {
  name         = "tf_provider_data_processing_feed"
  billing_mode = "PAY_PER_REQUEST"

  hash_key = "Provider"
  range_key = "LastVersion"

  attribute {
    name = "Provider"
    type = "S"
  }

  attribute {
    name = "LastVersion"
    type = "S"
  }

  # attribute {
  #   name = "Etag"
  #   type = "S"
  # }

  lifecycle {
    ignore_changes = [
      #billing_mode,
      read_capacity,
      write_capacity,
    ]

  }

  tags = local.tags
}
