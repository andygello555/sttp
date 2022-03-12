from django.urls import include, path
from rest_framework import routers

from blog_api_server.api import views

router = routers.DefaultRouter(trailing_slash=False)
router.register(r'topics', views.TopicViewSet)
router.register(r'blogs', views.BlogViewSet)
router.register(r'comments', views.CommentViewSet)

urlpatterns = [
    path('', include(router.urls))
]
