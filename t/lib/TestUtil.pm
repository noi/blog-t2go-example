package TestUtil;
use strict;
use warnings;
use Exporter 'import';
use HTTP::Tiny;
use JSON::PP;

our @EXPORT_OK = qw(create_user get post);

my $base = $ENV{TEST_BASE_URL} or die 'TEST_BASE_URL is not set';
my $http = HTTP::Tiny->new;
my $json = JSON::PP->new->utf8;
my $seq  = 0;

sub create_user {
    return join '-', 'user', $$, ++$seq;
}

# get / post は { status => HTTP ステータス, body => デコードした JSON } を返す。
sub get {
    my ($user, $path) = @_;
    my $res = $http->get("$base$path", { headers => { 'X-User-Id' => $user } });
    return _decode($res);
}

sub post {
    my ($user, $path, $params) = @_;
    my $res = $http->post("$base$path", {
        headers => { 'X-User-Id' => $user, 'Content-Type' => 'application/json' },
        content => $json->encode($params),
    });
    return _decode($res);
}

sub _decode {
    my $res  = shift;
    my $body = eval { $json->decode($res->{content}) };
    return { status => $res->{status}, body => $body };
}

1;
