import { useState, useRef, useCallback } from "react";
import Cropper, { ReactCropperElement } from "react-cropper";
import "cropperjs/dist/cropper.css";
import type { MediaItem } from "@/api/media";
import { formatUploadError, uploadMediaTracked } from "@/lib/mediaUploadTracked";
import {
  ZoomIn,
  ZoomOut,
  RotateCcw,
  RotateCw,
  RefreshCw,
} from "lucide-react";

interface ImageCropUploadProps {
  onUpload: (item: MediaItem) => void;
  aspectRatio?: number;
  currentImageUrl?: string;
}

export default function ImageCropUpload({
  onUpload,
  aspectRatio,
  currentImageUrl,
}: ImageCropUploadProps) {
  const [imageSrc, setImageSrc] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCropper, setShowCropper] = useState(false);
  const cropperRef = useRef<ReactCropperElement>(null);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith("image/")) {
      setError("请选择图片文件");
      return;
    }

    setError(null);
    const reader = new FileReader();
    reader.onload = () => {
      setImageSrc(reader.result as string);
      setShowCropper(true);
    };
    reader.readAsDataURL(file);
    // Reset input so same file can be selected again
    e.target.value = "";
  };

  const getCroppedBlob = useCallback((): Promise<Blob> => {
    return new Promise((resolve, reject) => {
      const cropper = cropperRef.current?.cropper;
      if (!cropper) {
        reject(new Error("Cropper not initialized"));
        return;
      }
      const canvas = cropper.getCroppedCanvas();
      if (!canvas) {
        reject(new Error("Failed to get cropped canvas"));
        return;
      }
      canvas.toBlob(
        (blob) => {
          if (blob) resolve(blob);
          else reject(new Error("Canvas toBlob failed"));
        },
        "image/jpeg",
        0.9
      );
    });
  }, []);

  const handleCropConfirm = async () => {
    if (!imageSrc) return;

    setUploading(true);
    setError(null);

    try {
      const croppedBlob = await getCroppedBlob();
      const filename = `cropped-${Date.now()}.jpg`;
      const item = await uploadMediaTracked(croppedBlob, { filename });
      onUpload(item);
      setShowCropper(false);
      setImageSrc(null);
    } catch (err) {
      setError(formatUploadError(err, "上传失败"));
    } finally {
      setUploading(false);
    }
  };

  const handleDirectUpload = async () => {
    if (!imageSrc) return;

    setUploading(true);
    setError(null);

    try {
      const response = await fetch(imageSrc);
      const blob = await response.blob();
      const filename = `upload-${Date.now()}.jpg`;
      const item = await uploadMediaTracked(blob, { filename });
      onUpload(item);
      setShowCropper(false);
      setImageSrc(null);
    } catch (err) {
      setError(formatUploadError(err, "上传失败"));
    } finally {
      setUploading(false);
    }
  };

  const handleCancel = () => {
    setShowCropper(false);
    setImageSrc(null);
    setError(null);
  };

  const handleZoom = (delta: number) => {
    cropperRef.current?.cropper.zoom(delta);
  };

  const handleRotate = (degree: number) => {
    cropperRef.current?.cropper.rotate(degree);
  };

  const handleReset = () => {
    cropperRef.current?.cropper.reset();
  };

  return (
    <div>
      {/* Current image preview */}
      {currentImageUrl && !showCropper && (
        <div className="mb-3">
          <img
            src={currentImageUrl}
            alt=""
            className="max-h-24 rounded border border-gray-200 object-contain"
          />
        </div>
      )}

      {/* File input */}
      {!showCropper && (
        <label className="inline-flex items-center px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 cursor-pointer">
          选择图片
          <input
            type="file"
            accept="image/*"
            onChange={handleFileSelect}
            className="hidden"
          />
        </label>
      )}

      {/* Crop modal */}
      {showCropper && imageSrc && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white rounded-xl shadow-xl w-[90vw] max-w-2xl">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">裁剪图片</h3>
            </div>

            <div className="h-96 bg-gray-900">
              <Cropper
                ref={cropperRef}
                src={imageSrc}
                style={{ height: "100%", width: "100%" }}
                initialAspectRatio={aspectRatio ?? 16 / 9}
                aspectRatio={aspectRatio ?? NaN}
                viewMode={1}
                guides={true}
                cropBoxMovable={true}
                cropBoxResizable={true}
                autoCropArea={0.8}
                zoomOnWheel={true}
                background={true}
                responsive={true}
              />
            </div>

            {/* Toolbar */}
            <div className="px-6 py-3 flex items-center justify-center gap-2 border-t border-gray-100">
              <button
                type="button"
                onClick={() => handleZoom(-0.1)}
                className="p-2 rounded-md hover:bg-gray-100 text-gray-600"
                title="缩小"
              >
                <ZoomOut size={18} />
              </button>
              <button
                type="button"
                onClick={() => handleZoom(0.1)}
                className="p-2 rounded-md hover:bg-gray-100 text-gray-600"
                title="放大"
              >
                <ZoomIn size={18} />
              </button>
              <div className="w-px h-5 bg-gray-300 mx-1" />
              <button
                type="button"
                onClick={() => handleRotate(-90)}
                className="p-2 rounded-md hover:bg-gray-100 text-gray-600"
                title="左旋 90°"
              >
                <RotateCcw size={18} />
              </button>
              <button
                type="button"
                onClick={() => handleRotate(90)}
                className="p-2 rounded-md hover:bg-gray-100 text-gray-600"
                title="右旋 90°"
              >
                <RotateCw size={18} />
              </button>
              <div className="w-px h-5 bg-gray-300 mx-1" />
              <button
                type="button"
                onClick={handleReset}
                className="p-2 rounded-md hover:bg-gray-100 text-gray-600"
                title="重置"
              >
                <RefreshCw size={18} />
              </button>
            </div>

            {error && (
              <div className="px-6 py-2">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <div className="px-6 py-4 border-t border-gray-200 flex items-center justify-end gap-3">
              <button
                type="button"
                onClick={handleCancel}
                disabled={uploading}
                className="px-4 py-2 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
              >
                取消
              </button>
              <button
                type="button"
                onClick={handleDirectUpload}
                disabled={uploading}
                className="px-4 py-2 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
              >
                {uploading ? "上传中..." : "不裁剪直接上传"}
              </button>
              <button
                type="button"
                onClick={handleCropConfirm}
                disabled={uploading}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {uploading ? "上传中..." : "裁剪并上传"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
