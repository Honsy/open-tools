import Home from "./pages/Home.vue";
import Category from "./pages/Category.vue";
import Search from "./pages/Search.vue";
import Article from "./pages/Article.vue";
import Submit from "./pages/Submit.vue";
import Tool from "./pages/Tool.vue";
import AdminLogin from "./pages/admin/Login.vue";
import AdminLayout from "./pages/admin/Layout.vue";
import AdminReview from "./pages/admin/Review.vue";
import AdminLinks from "./pages/admin/Links.vue";
import AdminCategories from "./pages/admin/Categories.vue";
import AdminArticles from "./pages/admin/Articles.vue";
import AdminTags from "./pages/admin/Tags.vue";

export const routes = [
  { path: "/", component: Home },
  { path: "/c/:slug", component: Category },
  { path: "/search", component: Search },
  { path: "/a/:id", component: Article },
  { path: "/submit", component: Submit },
  { path: "/tools/:slug", component: Tool },
  { path: "/admin/login", component: AdminLogin },
  {
    path: "/admin",
    component: AdminLayout,
    children: [
      { path: "", component: AdminReview },
      { path: "links", component: AdminLinks },
      { path: "categories", component: AdminCategories },
      { path: "articles", component: AdminArticles },
      { path: "tags", component: AdminTags },
    ],
  },
];
