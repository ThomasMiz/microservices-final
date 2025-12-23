use std::num::NonZeroU32;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct PageRequest {
    page_number: u32,
    page_size: NonZeroU32,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PageResult<E> {
    request: PageRequest,
    elements: Vec<E>,
    total_elements: u64,
}

impl PageRequest {
    pub fn new(page_number: u32, page_size: NonZeroU32) -> Self {
        Self {
            page_number,
            page_size,
        }
    }

    pub fn page_number(&self) -> u32 {
        self.page_number
    }

    pub fn page_size(&self) -> NonZeroU32 {
        self.page_size
    }

    pub fn page_size_u32(&self) -> u32 {
        self.page_size.get()
    }

    pub fn offset(&self) -> u64 {
        (self.page_number as u64) * (self.page_size.get() as u64)
    }
}

impl<E> PageResult<E> {
    pub fn new(request: PageRequest, elements: Vec<E>, total_elements: u64) -> Self {
        Self {
            request,
            elements,
            total_elements,
        }
    }

    pub fn request(&self) -> PageRequest {
        self.request
    }

    pub fn elements(&self) -> &Vec<E> {
        &self.elements
    }

    pub fn total_elements(&self) -> u64 {
        self.total_elements
    }

    pub fn into_elements(self) -> Vec<E> {
        self.elements
    }
}

#[derive(sqlx::FromRow, Debug, Clone, Copy, PartialEq, Eq)]
pub struct PageCountEntity {
    pub count: i64, // Note: sqlx+postgres does not support u64 :-(
}
