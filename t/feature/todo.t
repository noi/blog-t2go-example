use strict;
use warnings;
use utf8;
use Test::More;
use Test::Deep;
use lib 't/lib';
use TestUtil qw(create_user get post);

subtest 'create' => sub {
    my $user = create_user();
    my $res  = post($user, '/api/todo/create', { title => '牛乳を買う' });
    is $res->{status}, 200;
    cmp_deeply $res->{body}, {
        todo => {
            id         => ignore(),
            title      => '牛乳を買う',
            done       => bool(0),
            created_at => ignore(),
        },
    };
};

subtest 'create without title' => sub {
    my $user = create_user();
    my $res  = post($user, '/api/todo/create', { title => '' });
    is $res->{status},        400;
    is $res->{body}{error},   'title is required';
};

subtest 'list' => sub {
    my $user = create_user();
    post($user, '/api/todo/create', { title => $_ }) for qw(a b c);
    my $other = create_user();
    post($other, '/api/todo/create', { title => 'other' });

    my $res = get($user, '/api/todo/list');
    cmp_deeply [ map { $_->{title} } @{ $res->{body}{todos} } ], bag(qw(a b c));
};

subtest 'done' => sub {
    my $user = create_user();
    my $todo = post($user, '/api/todo/create', { title => 'done me' })->{body}{todo};

    my $res = post($user, '/api/todo/done', { id => $todo->{id} });
    cmp_deeply $res->{body}{todo}, superhashof({
        id    => $todo->{id},
        title => ignore
        done  => 1,
    });
};

subtest 'done by other user' => sub {
    my $user  = create_user();
    my $todo  = post($user, '/api/todo/create', { title => 'mine' })->{body}{todo};
    my $other = create_user();

    my $res = post($other, '/api/todo/done', { id => $todo->{id} });
    is $res->{status}, 404;
};

subtest 'archive' => sub {
    my $user = create_user();
    my $res  = post($user, '/api/todo/archive', { id => 1 });
    is $res->{status}, 404;
};

done_testing;
