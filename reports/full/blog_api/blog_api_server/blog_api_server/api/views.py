from django.db.models import F

from rest_framework import viewsets, status
from rest_framework.decorators import action
from rest_framework.permissions import AllowAny
from rest_framework.request import Request
from rest_framework.response import Response

from blog_api_server.api.models import Topic, Blog, Comment
from blog_api_server.api.serialisers import TopicSerialiser, BlogSerialiser, CommentSerialiser


class TopicViewSet(viewsets.ModelViewSet):
    queryset = Topic.objects.all().order_by('name')
    serializer_class = TopicSerialiser
    permission_classes = [AllowAny]

    @action(detail=False)
    def top(self, request: Request) -> Response:
        return Response(
            TopicSerialiser(sorted(self.get_queryset(), key=lambda t: t.blog_count, reverse=True)[:10], many=True).data,
            status=status.HTTP_200_OK
        )

    @action(detail=True)
    def blogs(self, request: Request, pk=None) -> Response:
        return Response(
            BlogSerialiser(self.get_object().blogs.all().order_by('name'), many=True, context={'request': request}).data,
            status=status.HTTP_200_OK
        )

class BlogViewSet(viewsets.ModelViewSet):
    queryset = Blog.objects.all().order_by('name')
    serializer_class = BlogSerialiser
    permission_classes = [AllowAny]

    @action(detail=True)
    def comments(self, request: Request, pk=None) -> Response:
        return Response(
            CommentSerialiser(self.get_object().comments.all(), many=True, context={'request': request}).data,
            status=status.HTTP_200_OK
        )

class CommentViewSet(viewsets.ModelViewSet):
    queryset = Comment.objects.all().order_by('commenter')
    serializer_class = CommentSerialiser
    permission_classes = [AllowAny]

