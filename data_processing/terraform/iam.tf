### Lambda IAM
resource "aws_iam_role" "tf_provider_data_processing_lambda" {
  name               = "tf_provider_data_processing_lambda"
  assume_role_policy = data.aws_iam_policy_document.tf_provider_data_processing_lambda.json

  tags = local.tags
}

data "aws_iam_policy_document" "tf_provider_data_processing_lambda" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "tf_provider_data_processing_lambda_dynamodb" {
  statement {
    effect    = "Allow"
    actions   = ["dynamodb:*"]
    resources = ["arn:aws:dynamodb:::table/tf_provider_data_processing*"]
  }
}

resource "aws_iam_role_policy" "dynamodb" {
  name   = "tf_provider_data_processing_lambda_dynamodb"
  role   = aws_iam_role.tf_provider_data_processing_lambda.id
  policy = data.aws_iam_policy_document.tf_provider_data_processing_lambda_dynamodb.json
}

### Step Functions IAM
resource "aws_iam_role" "tf_provider_data_processing_stepfunction" {
  name               = "tf_provider_data_processing_stepfunction"
  assume_role_policy = data.aws_iam_policy_document.tf_provider_data_processing_stepfunction.json

  tags = local.tags
}

data "aws_iam_policy_document" "tf_provider_data_processing_stepfunction" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["states.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "xray" {
  role       = aws_iam_role.tf_provider_data_processing_stepfunction.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
}
