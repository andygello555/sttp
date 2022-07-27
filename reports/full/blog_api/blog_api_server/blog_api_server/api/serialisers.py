from rest_framework import serializers

from blog_api_server.api.models import Topic, Blog, Comment


class TopicSerialiser(serializers.ModelSerializer):
    blog_count = serializers.ReadOnlyField()

    class Meta:
        model = Topic
        fields = '__all__'

class BlogSerialiser(serializers.ModelSerializer):
    topic = serializers.HyperlinkedRelatedField(view_name='topic-detail', read_only=True)
    topic_id = serializers.PrimaryKeyRelatedField(queryset=Topic.objects.all(), source='topic', write_only=True)

    class Meta:
        model = Blog
        fields = '__all__'

class CommentSerialiser(serializers.ModelSerializer):
    blog = serializers.HyperlinkedRelatedField(view_name='blog-detail', read_only=True)
    blog_id = serializers.PrimaryKeyRelatedField(queryset=Blog.objects.all(), source='blog', write_only=True)

    class Meta:
        model = Comment
        fields = '__all__'
