data "spacelift_ips" "ips" {}

resource "aws_security_group_rule" "allow_spacelift" {
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = data.spacelift_ips.ips.cidrs
  security_group_id = aws_security_group.example.id
  description       = "Allow Spacelift IPs"
}
