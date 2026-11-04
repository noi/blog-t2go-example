use strict;
use warnings;
use utf8;
use Test::More;
use Test::Deep;
use lib 't/lib';
use TestUtil qw(create_user api_get api_post);

subtest 'create' => sub {
    my $user = create_user();
    my $res  = api_post($user, '/api/todo/create', { title => '牛乳を買う' });
    cmp_deeply $res, {
        status => { ok => 1 },
        todo   => {
            id         => ignore(),
            title      => '牛乳を買う',
            done       => bool(0),
            created_at => ignore(),
        },
    };
};

subtest 'create without title' => sub {
    my $user = create_user();
    my $res  = api_post($user, '/api/todo/create', { title => '' });
    is $res->{status}{ok},    0;
    is $res->{status}{error}, 'title is required';
};

subtest 'list' => sub {
    my $user = create_user();
    api_post($user, '/api/todo/create', { title => $_ }) for qw(a b c);
    my $other = create_user();
    api_post($other, '/api/todo/create', { title => 'other' });

    my $res = api_get($user, '/api/todo/list');
    cmp_deeply [ map { $_->{title} } @{ $res->{todos} } ], bag(qw(a b c));
};

subtest 'done' => sub {
    my $user = create_user();
    my $todo = api_post($user, '/api/todo/create', { title => 'done me' })->{todo};

    my $res = api_post($user, '/api/todo/done', { id => $todo->{id} });
    cmp_deeply $res->{todo}, superhashof({
        id    => $todo->{id},
        title => ignore
        done  => 1,
    });
};

subtest 'done by other user' => sub {
    my $user  = create_user();
    my $todo  = api_post($user, '/api/todo/create', { title => 'mine' })->{todo};
    my $other = create_user();

    my $res = api_post($other, '/api/todo/done', { id => $todo->{id} });
    is $res->{status}{ok}, 0;
};

subtest 'archive' => sub {
    my $user = create_user();
    my $res  = api_post($user, '/api/todo/archive', { id => 1 });
    is $res->{http_status}, 404;
};

done_testing;
