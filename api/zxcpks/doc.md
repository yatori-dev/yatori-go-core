# 在线评测考试平台
## 接口说明
这个接口用于访问对应学校域名获得对应学校相关数据，其中domain参数添加的是对应学校域名
```bash
curl 'https://swxy.haiqikeji.com/api/course/selectdomain?domain=swxy.haiqikeji.com' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'accept-language: zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7' \
  -H 'cache-control: no-cache' \
  -b '__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qdda=3-1.1; _qddab=3-b2u746.mmrwmvxj' \
  -H 'pragma: no-cache' \
  -H 'priority: u=1, i' \
  -H 'referer: https://swxy.haiqikeji.com/' \
  -H 'sec-ch-ua: "Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36'
```
返回数据：
```azure
{
    "code": 200,
    "msg": "success",
    "data": {
        "id": 15, //学校id
        "name": "中南林业科技大学涉外学院",
        "nameEn": "Swan college,Central South University of Forestry and Technology",
        "ident": "zswxy",
        "area": "[]",
        "province": 18,
        "city": 220,
        "region": 0,
        "badge": "https://s2.yinghuaonline.com/upfiles/5eoQSGytIPLF8zyf6BfM.png",
        "logo": "https://guozejiaoyu-1317857445.cos.ap-nanjing.myqcloud.com/2026/03/02/785cedf079764f86befe91f3d602da79.png",
        "address": "长沙市望城县丁字镇102省道",
        "website": "https://swxy.csuft.edu.cn/",
        "intro": "中南林业科技大学涉外学院（Swan college of Central South University of Forestry and Technology ）成立于2002年6月，是经湖南省人民政府批准设立，国家教育部确认，由中南林业科技大学与湖南旺湘科技产业投资有限公司合作举办，具有独立法人资格的全日制本科独立学院。\r\n学院位于现代化城市长沙市，总占地面积2750余亩（含实习、实训、科研、产业基地），校舍面积27万余平方米。建有教学楼、实验楼、图书馆、学生宿舍、食堂等教学、文娱、服务场所和校园网、多媒体教室、语音室、同声传译室、计算机房、电子阅览室、体操房、跆拳道室、乒乓球室等现代化教学文体设施，并共享中南林业科技大学部分教学资源。校园环境幽雅，教学设施先进，师资队伍精良，办学特色鲜明，具有良好的办学条件和发展前景。\r\n校本部，中南林业科技大学是湖南省人民政府和国家林业局重点建设高校，是湖南省五所高水平大学之一，湖南省属具有研究生推免资格的六所高校之一，亦是湖南省第一个拥有研究生院的省属高校，省部共建大学、省属重点大学、中西部高校基础能力建设工程（小211工程）、卓越农林人才教育培养计划、湖南省2011计划建设高校，全国本科一批招生。学校涵盖理、工、农、文、经、法、管、教、艺等九大学科门类，是具有博士后科研流动站、博士学位授予权和硕士生推免权、以林业科学为特色的综合型大学。[1] ",
        "allow": 1,
        "addTime": "2022-02-28 13:36:00",
        "createId": 1,
        "oldPlatformId": null,
        "useCourse": 1,
        "cooperate": 0,
        "sort": 85,
        "content": "<p>中南林业科技大学涉外学院（Swan college of Central South University of Forestry and Technology ）成立于2002年6月，是经湖南省人民政府批准设立，国家教育部确认，由中南林业科技大学与湖南旺湘科技产业投资有限公司合作举办，具有独立法人资格的全日制本科独立学院。 学院位于现代化城市长沙市，总占地面积2750余亩（含实习、实训、科研、产业基地），校舍面积27万余平方米。建有教学楼、实验楼、图书馆、学生宿舍、食堂等教学、文娱、服务场所和校园网、多媒体教室、语音室、同声传译室、计算机房、电子阅览室、体操房、跆拳道室、乒乓球室等现代化教学文体设施，并共享中南林业科技大学部分教学资源。校园环境幽雅，教学设施先进，师资队伍精良，办学特色鲜明，具有良好的办学条件和发展前景。 校本部，中南林业科技大学是湖南省人民政府和国家林业局重点建设高校，是湖南省五所高水平大学之一，湖南省属具有研究生推免资格的六所高校之一，亦是湖南省第一个拥有研究生院的省属高校，省部共建大学、省属重点大学、中西部高校基础能力建设工程（小211工程）、卓越农林人才教育培养计划、湖南省2011计划建设高校，全国本科一批招生。学校涵盖理、工、农、文、经、法、管、教、艺等九大学科门类，是具有博士后科研流动站、博士学位授予权和硕士生推免权、以林业科学为特色的综合型大学。[1]</p>",
        "banner": "https://bfbtmooc-1256316799.cos.ap-beijing.myqcloud.com/2026/03/06/2319ecd419de4643be63c26afe022bac.png",
        "contact": " 中南林业科技大学涉外学院——在线教学平台\r\n 地址 :湖南省长沙市芙蓉北路特1号\r\n 网络中心电话：0731-89814016\r\n 教务处电话：0731-89814002",
        "copyright": "",
        "map": "湖南省 长沙市",
        "domain": "swxy.haiqikeji.com",
        "domainType": 2,
        "skin": "default"
    },
    "count": null,
    "extra": {},
    "list": null
}
```