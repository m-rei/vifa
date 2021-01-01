initCarousel();
function initCarousel() {
    let carousels = document.querySelectorAll(".media-carousel");
    for (let carousel of carousels) {
        let left = carousel.querySelector(".left")
        let right = carousel.querySelector(".right")
        let images = carousel.querySelectorAll(".images img");
        let indicators = carousel.querySelectorAll(".indicators i");
        
        left.addEventListener("click", e => {
            // update active image
            let activeImgIdx = -1;
            for (let i = 0; i < images.length; i++) {
                if (images[i].classList.contains("active")) {
                    activeImgIdx = i;
                    break;
                }
            }
            if (activeImgIdx != -1) {
                let prev = activeImgIdx-1;
                if (prev < 0) {
                    prev = images.length + prev;
                }
                images[prev].classList.add("active");
                images[activeImgIdx].classList.remove("active");
            }
            // update active indicator
            let activeIndIdx = -1;
            for (let i = 0; i < indicators.length; i++) {
                if (indicators[i].classList.contains("active")) {
                    activeIndIdx = i;
                    break;
                }
            }
            if (activeIndIdx != -1) {
                let prev = activeIndIdx-1;
                if (prev < 0) {
                    prev = indicators.length + prev;
                }
                indicators[prev].classList.add("active");
                indicators[activeImgIdx].classList.remove("active");
            }
        });
        right.addEventListener("click", e => {
            // update active image
            let activeImgIdx = -1;
            for (let i = 0; i < images.length; i++) {
                if (images[i].classList.contains("active")) {
                    activeImgIdx = i;
                    break;
                }
            }
            if (activeImgIdx != -1) {
                let next = activeImgIdx+1;
                if (next >= images.length) {
                    next -= images.length
                }
                images[activeImgIdx].classList.remove("active");
                images[next].classList.add("active");
            }
            // update active indicator
            let activeIndIdx = -1;
            for (let i = 0; i < indicators.length; i++) {
                if (indicators[i].classList.contains("active")) {
                    activeIndIdx = i;
                    break;
                }
            }
            if (activeIndIdx != -1) {
                let next = activeIndIdx+1;
                if (next >= images.length) {
                    next -= images.length
                }
                indicators[activeImgIdx].classList.remove("active");
                indicators[next].classList.add("active");
            }
        });

        for (let indicator of indicators) {
            indicator.addEventListener("click", e => {
                let slideTo = indicator.dataset.slideTo;

                let activeIndIdx = -1;
                for (let i = 0; i < indicators.length; i++) {
                    if (indicators[i].classList.contains("active")) {
                        activeIndIdx = i;
                        break;
                    }
                }
                if (activeIndIdx != -1 && activeIndIdx != slideTo) {
                    indicators[activeIndIdx].classList.remove("active");
                    indicators[slideTo].classList.add("active");
                    images[activeIndIdx].classList.remove("active");
                    images[slideTo].classList.add("active");
                }
            });
        }
    }
}