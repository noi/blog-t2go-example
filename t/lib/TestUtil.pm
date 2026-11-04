package TestUtil;
use strict;
use warnings;
use Exporter 'import';
use HTTP::Tiny;
use JSON::PP;

our @EXPORT_OK = qw(create_user api_get api_post);

my $base = $ENV{TEST_BASE_URL} or die 'TEST_BASE_URL is not set';
my $http = HTTP::Tiny->new;
my $json = JSON::PP->new->utf8;
my $seq  = 0;

sub create_user {
    return join '-', 'user', $$, ++$seq;
}

sub api_get {
    my ($user, $path) = @_;
    my $res = $http->get("$base$path", { headers => { 'X-User-Id' => $user } });
    return _decode($res);
}

sub api_post {
    my ($user, $path, $params) = @_;
    my $res = $http->post("$base$path", {
        headers => { 'X-User-Id' => $user, 'Content-Type' => 'application/json' },
        content => $json->encode($params),
    });
    return _decode($res);
}

sub _decode {
    my $res = shift;
    return { http_status => $res->{status} } unless $res->{status} == 200;
    return $json->decode($res->{content});
}

1;
