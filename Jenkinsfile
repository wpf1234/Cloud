env.PROJ_DIR="${JENKINS_HOME}/workspace"        //jenkins  workspace
env.PROJ_URL="git@github.gree.com:180435/cloudservice.git"  // 项目的SSH地址
//env.PROJ_NAME="SoftwareManageSystem" // 项目名
env.LANGUAGE="golang" //基础镜像
env.TAGS="1.12.4-alpine3.9"     //基础镜像TAG
env.HARBOR="dimage.gree.com/devops"  //镜像仓库的URL
env.INAME="greeyun"  //制作的镜像名(自定义时不要使用大写字母)
//    String tag='latest' //制作的镜像TAG
node {


        stage('Get Code') {
            dir(path: "${PROJ_DIR}/$JOB_NAME/") {    //指定git拉取代码目录
                git "${PROJ_URL}"
                tag = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()  // 获取 git  commitID
                }
        }

        withEnv(["IMAGE_TAG=${tag}"]) {
            stage('Docker Build') {           //构建镜像,ssh到远程主机执行命令，该命令要用""，Dodckerfile顶格
            sh '''
                  ssh root@10.2.47.12 "rm -rf /opt/$JOB_NAME"
                  scp -r ${PROJ_DIR}/$JOB_NAME root@10.2.47.12:/opt/;
                  ssh root@10.2.47.12 "cd /opt/$JOB_NAME;
                  cat << EOF > Dockerfile
FROM ${LANGUAGE}:${TAGS} as builder
WORKDIR /go/src/$JOB_NAME
COPY . .
RUN go build -o $JOB_NAME

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/$JOB_NAME/$JOB_NAME .
CMD [\\"./$JOB_NAME\\"]
EOF
                  docker build -t ${HARBOR}/${INAME}:$IMAGE_TAG ."
              '''
           }
       stage('Image Push') {              //上传镜像到远程镜像仓库
                     sh '''
                     ssh root@10.2.47.12 "docker push ${HARBOR}/${INAME}:$IMAGE_TAG;
                     rm -rf /opt/$JOB_NAME;
                     "
                     '''
       }
       
             
        stage('Remove old container') {              //删除原有容器
            try{
                 sh '''
                 ssh centos@10.2.15.92 "docker rm -f $JOB_NAME"
                 '''
                }catch (error){
                }finally{
                    echo "remove old container success"
                }
        }

        stage('Image Pull') {              //从远程镜像仓库拉取镜像
            sh '''
            ssh centos@10.2.15.92 "docker pull ${HARBOR}/${INAME}:$IMAGE_TAG"
            '''
        }

        stage('Docker Run') {              //在服务器上起docker镜像
                     
            sh '''
            ssh centos@10.2.15.92 "docker run -i -d -p 11111:11111 -v /home/centos/marketplace/cloud-wpf/conf/application.yml:/root/conf/application.yml -v /home/centos/marketplace/cloud-wpf/log:/root/log --name=$JOB_NAME ${HARBOR}/${INAME}:${IMAGE_TAG}"
            '''
        }

      }
}
