package video

/*
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libswscale/swscale.h>
#include <libavutil/imgutils.h>

//包含header的目录
#cgo CFLAGS: -IC:/ffmpeg_3.4.2/include

#cgo LDFLAGS:  -lavutil -lavcodec -lavformat -lswscale

void saveFrame(AVFrame*,int,int,int);

int start(char* path,char* dst)
{

    //av_register_all();

    AVFormatContext *pFormatCtx = NULL;

    // Open video file.
    if (avformat_open_input(&pFormatCtx, path, NULL, NULL) != 0)
    {
        return -1; // Couldn't open file.
    }

    // Retrieve stream information.
    if (avformat_find_stream_info(pFormatCtx, NULL) < 0)
    {
        return -1; // Couldn't find stream information.
    }

    // Dump information about file onto standard error.
    //av_dump_format(pFormatCtx, 0, path, 0);

    // Find the first video stream.
    int videoStream;
    videoStream = -1;
    int i;
    for ( i = 0; i < pFormatCtx->nb_streams; i++)
    {
        if (pFormatCtx->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_VIDEO)
        {
            videoStream = i;
            break;
        }
    }
    if (videoStream == -1)
    {
        return -1; // Didn't find a video stream.
    }

    // Find the decoder for the video stream.
    AVCodec *pCodec;
    pCodec = avcodec_find_decoder(pFormatCtx->streams[videoStream]->codecpar->codec_id);
    if (pCodec == NULL)
    {
        fprintf(stderr, "Unsupported codec!\n");
        return -1; // Codec not found.
    }
    // Copy context.
    AVCodecContext *pCodecCtx;
    pCodecCtx = avcodec_alloc_context3(pCodec);
    if (avcodec_parameters_to_context(pCodecCtx, pFormatCtx->streams[videoStream]->codecpar) < 0)
    {
        fprintf(stderr, "Couldn't copy codec context");
        return -1; // Error copying codec context.
    }

    // Open codec.
    if (avcodec_open2(pCodecCtx, pCodec, NULL) < 0)
    {
        return -1; // Could not open codec.
    }

    AVFrame *pFrame = NULL;

    // Allocate video frame.
    pFrame = av_frame_alloc();

    AVPacket packet;

    int width = pFormatCtx->streams[videoStream]->codecpar->width;
    int height = pFormatCtx->streams[videoStream]->codecpar->height;
    int dstWidth,dstHeight;
    dstWidth = width*10;
    dstHeight=height*10;

    AVFrame *pDstFrame;
    pDstFrame = av_frame_alloc();
    pDstFrame->width = dstWidth;
    pDstFrame->height = dstHeight;
    pDstFrame->format = AV_PIX_FMT_YUVJ420P;
    int ret = av_frame_get_buffer(pDstFrame, 32);
    if (ret < 0) {
        fprintf(stderr, "Could not allocate the video frame data\n");
        exit(1);
    }

    int64_t duration;
    int timeStep,seek;
    seek = 1;
    duration = (pFormatCtx->duration+ (pFormatCtx->duration <= INT64_MAX - 5000 ? 5000 : 0))/AV_TIME_BASE;
    timeStep = (duration-seek)/100;

    for(i=0;i<100;i++){
        av_seek_frame(pFormatCtx,-1,((int64_t)(pFormatCtx->start_time + seek) * AV_TIME_BASE),AVSEEK_FLAG_FRAME);
        readFrame(pFormatCtx,pCodecCtx,&packet,pFrame,videoStream,i,dst,pDstFrame);
        seek += timeStep;
    }
    char dstProcess[1024];
    sprintf(dstProcess, "%sprocess.jpg",dst);

    saveDstFrame(dstProcess,pDstFrame);

    // Read frames and save .

    av_frame_free(&pFrame);

    // Close the codecs.
    avcodec_close(pCodecCtx);

    // Close the video file.
    avformat_close_input(&pFormatCtx);

    return 0;
}


int readFrame(AVFormatContext *pFormatCtx,AVCodecContext *pCodecCtx,AVPacket *pPacket,AVFrame *pFrame,int videoStream,int i,char* dst,AVFrame *pDstFrame){

    while (av_read_frame(pFormatCtx, pPacket) >= 0) {
        // Is this a packet from the video stream?
        if (pPacket->stream_index == videoStream && pPacket->flags == 1) {

            // Decode video frame
            avcodec_send_packet(pCodecCtx, pPacket);
            avcodec_receive_frame(pCodecCtx, pFrame);

            if(pFrame->width > 0){
                // Save the frame to disk.

                if(i == 0){
                    char tmp[1024];
                    sprintf(tmp, "%sthumb.jpg",dst);
                    img_saveSinge(pFrame,tmp);

                    // save first img as thumb
                }
                writeDstFrame(pDstFrame,pFrame,i);

                av_packet_unref(pPacket);
                break;
            }else{
                av_packet_unref(pPacket);
                continue;
            }
        }

        // Free the packet that was allocated by av_read_frame.
        av_packet_unref(pPacket);
    }
    // Free the packet that was allocated by av_read_frame.
    av_packet_unref(pPacket);
}

int writeDstFrame(AVFrame *pDstFrame,AVFrame *pFrame,int pos){
    int width = pFrame->width;
    int height = pFrame->height;
    int i,dstWidth,dstHeight,posX,posY;
    dstWidth = pDstFrame->width;
    dstHeight=pDstFrame->height;
    posX = pos % 10;
    posY = pos / 10;
    int y;
    for (y = 0; y < height; y++) {
        memcpy(pDstFrame->data[0] + y*dstWidth + posX * width + posY * height * dstWidth, pFrame->data[0] + y*width, width);
    }
    // Cb and Cr
    for (y = 0; y < height/2; y++) {
        memcpy(pDstFrame->data[1] + y*dstWidth/2 + posX * width/2 + posY * height * dstWidth / 4, pFrame->data[1] + y*width/2, width/2);
        memcpy(pDstFrame->data[2] + y*dstWidth/2 + posX * width/2 + posY * height * dstWidth / 4, pFrame->data[2] + y*width/2, width/2);
    }
    pDstFrame->pts = 1;
}

int saveDstFrame(char *out_filename,AVFrame *pFrame){
    int width = pFrame->width;
    int height = pFrame->height;

    AVCodecContext *pCodeCtx = NULL;
    AVFormatContext *pFormatCtx = avformat_alloc_context();
    // 设置输出文件格式
    pFormatCtx->oformat = av_guess_format("mjpeg", NULL, NULL);
    // 创建并初始化输出AVIOContext
    if (avio_open(&pFormatCtx->pb, out_filename, AVIO_FLAG_READ_WRITE) < 0)
    {
        printf("Couldn't open output file.");
        return -1;
    }
    // 构建一个新stream
    AVStream *pAVStream = avformat_new_stream(pFormatCtx, 0);
    if (pAVStream == NULL)
    {
        return -1;
    }
    AVCodecParameters *parameters = pAVStream->codecpar;
    parameters->codec_id = pFormatCtx->oformat->video_codec;
    parameters->codec_type = AVMEDIA_TYPE_VIDEO;
    parameters->format = AV_PIX_FMT_YUVJ420P;
    parameters->width = width;
    parameters->height = height;
    AVCodec *pCodec = avcodec_find_encoder(pAVStream->codecpar->codec_id);
    if (!pCodec)
    {
        printf("Could not find encoder\n");
        return -1;
    }
    pCodeCtx = avcodec_alloc_context3(pCodec);
    if (!pCodeCtx)
    {
        fprintf(stderr, "Could not allocate video codec context\n");
        exit(1);
    }
    if ((avcodec_parameters_to_context(pCodeCtx, pAVStream->codecpar)) < 0)
    {
        fprintf(stderr, "Failed to copy %s codec parameters to decoder context\n",
                av_get_media_type_string(AVMEDIA_TYPE_VIDEO));
        return -1;
    }
    AVRational Rat;
    Rat.num = 1;
    Rat.den = 25;
    pCodeCtx->time_base = Rat;
    if (avcodec_open2(pCodeCtx, pCodec, NULL) < 0)
    {
        printf("Could not open codec.");
        return -1;
    }
    int ret = avformat_write_header(pFormatCtx, NULL);
    if (ret < 0)
    {
        printf("write_header fail\n");
        return -1;
    }
    int64_t y_size = width * height;
    //Encode
    // 给AVPacket分配足够大的空间
    AVPacket pkt;
    av_new_packet(&pkt, y_size * 3);
    // 编码数据




    ret = avcodec_send_frame(pCodeCtx, pFrame);
    if (ret < 0)
    {
        printf("Could not avcodec_send_frame.");
        return -1;
    }

    av_frame_free(&pFrame);

    // 得到编码后数据

    ret = avcodec_receive_packet(pCodeCtx, &pkt);
    if (ret < 0)
    {
        printf("Could not avcodec_receive_packet");
        return -1;
    }
    ret = av_write_frame(pFormatCtx, &pkt);
    if (ret < 0)
    {
        printf("Could not av_write_frame");
        return -1;
    }

    av_packet_unref(&pkt);
    //Write Trailer
    av_write_trailer(pFormatCtx);
    avcodec_close(pCodeCtx);
    avio_close(pFormatCtx->pb);
    avformat_free_context(pFormatCtx);
    return 0;
}

int img_saveSinge(AVFrame *pFrame, char *out_filename)
{ //编码保存图片

    int width = pFrame->width;
    int height = pFrame->height;
    AVCodecContext *pCodeCtx = NULL;
    AVFormatContext *pFormatCtx = avformat_alloc_context();
    // 设置输出文件格式
    pFormatCtx->oformat = av_guess_format("mjpeg", NULL, NULL);

    // 创建并初始化输出AVIOContext
    if (avio_open(&pFormatCtx->pb, out_filename, AVIO_FLAG_READ_WRITE) < 0)
    {
        printf("Couldn't open output file.");
        return -1;
    }

    // 构建一个新stream
    AVStream *pAVStream = avformat_new_stream(pFormatCtx, 0);
    if (pAVStream == NULL)
    {
        return -1;
    }

    AVCodecParameters *parameters = pAVStream->codecpar;
    parameters->codec_id = pFormatCtx->oformat->video_codec;
    parameters->codec_type = AVMEDIA_TYPE_VIDEO;
    parameters->format = AV_PIX_FMT_YUVJ420P;
    parameters->width = pFrame->width;
    parameters->height = pFrame->height;

    AVCodec *pCodec = avcodec_find_encoder(pAVStream->codecpar->codec_id);

    if (!pCodec)
    {
        printf("Could not find encoder\n");
        return -1;
    }

    pCodeCtx = avcodec_alloc_context3(pCodec);
    if (!pCodeCtx)
    {
        fprintf(stderr, "Could not allocate video codec context\n");
        exit(1);
    }

    if ((avcodec_parameters_to_context(pCodeCtx, pAVStream->codecpar)) < 0)
    {
        fprintf(stderr, "Failed to copy %s codec parameters to decoder context\n",
                av_get_media_type_string(AVMEDIA_TYPE_VIDEO));
        return -1;
    }
    AVRational Rat;
    Rat.num = 1;
    Rat.den = 25;
    pCodeCtx->time_base = Rat;

    if (avcodec_open2(pCodeCtx, pCodec, NULL) < 0)
    {
        printf("Could not open codec.");
        return -1;
    }

    int ret = avformat_write_header(pFormatCtx, NULL);
    if (ret < 0)
    {
        printf("write_header fail\n");
        return -1;
    }

    int y_size = width * height;

    //Encode
    // 给AVPacket分配足够大的空间
    AVPacket pkt;
    av_new_packet(&pkt, y_size * 3);

    // 编码数据
    ret = avcodec_send_frame(pCodeCtx, pFrame);
    if (ret < 0)
    {
        printf("Could not avcodec_send_frame.");
        return -1;
    }

    // 得到编码后数据
    ret = avcodec_receive_packet(pCodeCtx, &pkt);
    if (ret < 0)
    {
        printf("Could not avcodec_receive_packet");
        return -1;
    }

    ret = av_write_frame(pFormatCtx, &pkt);

    if (ret < 0)
    {
        printf("Could not av_write_frame");
        return -1;
    }

    av_packet_unref(&pkt);

    //Write Trailer
    av_write_trailer(pFormatCtx);

    avcodec_close(pCodeCtx);
    avio_close(pFormatCtx->pb);
    avformat_free_context(pFormatCtx);

    return 0;
}


*/
import "C"
import (
	// "fmt"
	"unsafe"
)

// Deal video with dll
func Deal(file string, dstFile string) {
	//fmt.Println("start")
	input := C.CString(file)
	dst := C.CString(dstFile)
	C.start(input, dst)
	C.free(unsafe.Pointer(input))
	C.free(unsafe.Pointer(dst))

}

// func mergeImgs1(imgs []*bytes.Buffer, w, h int) *image.RGBA {
// 	target := image.NewRGBA(image.Rect(0, 0, w*10, h*10))
// 	for i, img := range imgs {
// 		decodedImg,_ := jpeg.Decode(img)
// 		position := decodedImg.Bounds().Add(image.Pt((i%10)*w, (i/10)*h))
// 		draw.Draw(target, position, decodedImg, decodedImg.Bounds().Min, draw.Src)
// 	}
// 	return target
// }
