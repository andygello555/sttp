import uuid

from django.db import models

# Create your models here.
class Topic(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4)
    name = models.CharField(max_length=64)

    @property
    def blog_count(self) -> int:
        return self.blogs.count()

class Blog(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4)
    name = models.CharField(max_length=256)
    body = models.TextField()
    author = models.CharField(max_length=128)
    topic = models.ForeignKey(Topic, on_delete=models.DO_NOTHING, related_name='blogs')
    tags = models.JSONField(default=list)

class Comment(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4)
    body = models.TextField()
    commenter = models.CharField(max_length=128)
    blog = models.ForeignKey(Blog, on_delete=models.DO_NOTHING, related_name='comments')
